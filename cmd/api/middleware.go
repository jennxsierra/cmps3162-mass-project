package main

import (
	"expvar"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// captures the status code for metrics purposes
type metricsResponseWriter struct {
	wrapped       http.ResponseWriter
	statusCode    int
	headerWritten bool
}

// creates a new instance of metricsResponseWriter
func newMetricsResponseWriter(w http.ResponseWriter) *metricsResponseWriter {
	return &metricsResponseWriter{
		wrapped:    w,
		statusCode: http.StatusOK,
	}
}

// implement the http.ResponseWriter interface for metricsResponseWriter
func (w *metricsResponseWriter) Header() http.Header {
	return w.wrapped.Header()
}

// capture the status code and mark that the header has been written
func (w *metricsResponseWriter) WriteHeader(statusCode int) {
	if w.headerWritten {
		return
	}
	w.statusCode = statusCode
	w.headerWritten = true
	w.wrapped.WriteHeader(statusCode)
}

// sets the status code to 200 OK
// if Write() is called without WriteHeader() being called first
func (w *metricsResponseWriter) Write(b []byte) (int, error) {
	if !w.headerWritten {
		w.WriteHeader(http.StatusOK)
	}
	return w.wrapped.Write(b)
}

// allows access to the underlying http.ResponseWriter
// for compatibility with http.HandlerFunc
func (w *metricsResponseWriter) Unwrap() http.ResponseWriter {
	return w.wrapped
}

// used for single numeric counter metrics (e.g. total requests received)
func getOrCreateExpvarInt(name string) *expvar.Int {
	metric := expvar.Get(name)
	if metric != nil {
		if counter, ok := metric.(*expvar.Int); ok {
			return counter
		}
	}
	return expvar.NewInt(name)
}

// used for map-based metrics (e.g. total responses sent by status code)
func getOrCreateExpvarMap(name string) *expvar.Map {
	metric := expvar.Get(name)
	if metric != nil {
		if counterMap, ok := metric.(*expvar.Map); ok {
			return counterMap
		}
	}
	return expvar.NewMap(name)
}

// tracks the total number of requests received, total responses sent,
// total processing time, and total responses sent by status code
func (a *applicationDependencies) metrics(next http.Handler) http.Handler {
	totalRequestsReceived := getOrCreateExpvarInt("total_requests_received")
	totalResponsesSent := getOrCreateExpvarInt("total_responses_sent")
	totalProcessingTimeMicroseconds := getOrCreateExpvarInt("total_processing_time_μs")
	totalResponsesSentByStatus := getOrCreateExpvarMap("total_responses_sent_by_status")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		totalRequestsReceived.Add(1)

		mw := newMetricsResponseWriter(w)
		next.ServeHTTP(mw, r)

		totalResponsesSent.Add(1)
		processingTimeMicroseconds := time.Since(start).Microseconds()
		totalProcessingTimeMicroseconds.Add(processingTimeMicroseconds)
		totalResponsesSentByStatus.Add(strconv.Itoa(mw.statusCode), 1)
	})
}

func (a *applicationDependencies) enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// indicates that responses may vary based on headers
		w.Header().Add("Vary", "Origin")
		w.Header().Add("Vary", "Access-Control-Request-Method")

		// get the value of the Origin header from the incoming request
		origin := r.Header.Get("Origin")

		if origin != "" {
			// check if the origin is in the trusted origins list
			for i := range a.config.cors.trustedOrigins {
				if origin == a.config.cors.trustedOrigins[i] {
					// allowing cross-origin requests for trusted origins
					w.Header().Set("Access-Control-Allow-Origin", origin)

					// handle CORS preflight requests
					if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
						// indicate allowed methods and headers for preflight requests
						w.Header().Set("Access-Control-Allow-Methods", "OPTIONS, PUT, PATCH, DELETE")
						w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")

						// respond to preflight requests with 200 OK and return early
						w.WriteHeader(http.StatusOK)
						return
					}

					break
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}

// recovers from any panics and sends a 500 Internal Server Error response
func (a *applicationDependencies) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// defer will be called when the stack unwinds
		defer func() {
			// recover() checks for panics
			err := recover()
			if err != nil {
				w.Header().Set("Connection", "close")
				a.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// limit the number of requests a client can make in a given time period
func (a *applicationDependencies) rateLimit(next http.Handler) http.Handler {
	// Define a rate limiter struct
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time // remove map entries that are stale
	}

	var mu sync.Mutex                      // use to synchronize the map
	var clients = make(map[string]*client) // the actual map

	// A goroutine to remove stale entries from the map
	go func() {
		for {
			time.Sleep(time.Minute)
			mu.Lock() // begin cleanup
			// delete any entry not seen in three minutes
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}
			mu.Unlock() // finish clean up
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// conditionally apply the rate limiter if enabled in the config
		if a.config.limiter.enabled {
			// get the IP address
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				a.serverErrorResponse(w, r, err)
				return
			}

			mu.Lock() // exclusive access to the map

			// check if ip address already in map, if not add it
			_, found := clients[ip]
			if !found {
				clients[ip] = &client{limiter: rate.NewLimiter(
					rate.Limit(a.config.limiter.rps),
					a.config.limiter.burst)}
			}

			// Update the last seen for the client
			clients[ip].lastSeen = time.Now()

			// Check the rate limit status
			if !clients[ip].limiter.Allow() {
				mu.Unlock() // no longer need exclusive access to the map
				a.rateLimitExceededResponse(w, r)
				return
			}

			mu.Unlock() // others are free to get exclusive access to the map
		}
		next.ServeHTTP(w, r)
	})
}

// logs HTTP requests with the request method, URL path, and the time it took to process the request
func (a *applicationDependencies) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Since(start)
		a.logger.Info("request received",
			"method", r.Method,
			"path", r.URL.Path,
			"duration", fmt.Sprintf("%.2fms", float64(duration.Microseconds())/1000.0),
		)
	})
}
