package main

import (
	"compress/gzip"
	"errors"
	"expvar"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/jennxsierra/mass-project/internal/data"
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

// tracks key request metrics:
// - total requests
// - requests per route
// - total error responses
// - average latency in milliseconds
func (a *applicationDependencies) metrics(next http.Handler) http.Handler {
	totalRequests := getOrCreateExpvarInt("total_requests")
	requestsPerRoute := getOrCreateExpvarMap("requests_per_route")
	totalErrorCount := getOrCreateExpvarInt("total_error_count")
	totalLatencyMicroseconds := getOrCreateExpvarInt("total_latency_microseconds")

	if expvar.Get("average_latency_ms") == nil {
		expvar.Publish("average_latency_ms", expvar.Func(func() any {
			requests := totalRequests.Value()
			if requests == 0 {
				return float64(0)
			}

			latencyMs := float64(totalLatencyMicroseconds.Value()) / 1000.0
			return latencyMs / float64(requests)
		}))
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		totalRequests.Add(1)
		requestsPerRoute.Add(r.URL.Path, 1)

		mw := newMetricsResponseWriter(w)
		next.ServeHTTP(mw, r)

		if mw.statusCode >= http.StatusBadRequest {
			totalErrorCount.Add(1)
		}

		totalLatencyMicroseconds.Add(time.Since(start).Microseconds())
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

// checks for a valid authentication token in the Authorization header of incoming requests
func (a *applicationDependencies) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// indicate that the response may vary based on the value of the Authorization header
		w.Header().Add("Vary", "Authorization")

		// get the value of the Authorization header from the incoming request
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" { // set anonymous user if no auth header provided
			r = a.contextSetUser(r, data.AnonymousUser)
			next.ServeHTTP(w, r)
			return
		}

		// check if the authHeader is in the correct format
		parts := strings.Fields(authHeader)
		if len(parts) != 2 || parts[0] != "Bearer" {
			a.invalidAuthenticationTokenResponse(w, r)
			return
		}

		tokenPlaintext := parts[1]
		if tokenPlaintext == "" {
			a.invalidAuthenticationTokenResponse(w, r)
			return
		}

		// retrieve the user associated with the token from the database
		user, err := a.models.User.GetForToken(data.ScopeAuthentication, tokenPlaintext)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				a.invalidAuthenticationTokenResponse(w, r)
			default:
				a.serverErrorResponse(w, r, err)
			}
			return
		}

		// set the user in the request context and call the next handler
		r = a.contextSetUser(r, user)
		next.ServeHTTP(w, r)
	})
}

// checks if the user is authenticated before allowing access to certain routes
func (a *applicationDependencies) requireAuthenticatedUser(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := a.contextGetUser(r)
		if user.IsAnonymous() {
			a.authenticationRequiredResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// checks if the user has an activated account before allowing access to certain routes
func (a *applicationDependencies) requireActivatedUser(next http.HandlerFunc) http.HandlerFunc {
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := a.contextGetUser(r)
		if !user.Activated {
			a.inactiveAccountResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})

	return a.requireAuthenticatedUser(fn)
}

// checks if the user has the necessary permissions before allowing access to certain routes
func (a *applicationDependencies) requirePermission(permissionCode string, next http.HandlerFunc) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		user := a.contextGetUser(r)

		permissions, err := a.models.Permission.GetAllForUser(user.ID)
		if err != nil {
			a.serverErrorResponse(w, r, err)
			return
		}

		if !permissions.Include(permissionCode) {
			a.notPermittedResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	}

	return a.requireActivatedUser(fn)
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

			reservation := clients[ip].limiter.Reserve()
			if !reservation.OK() {
				mu.Unlock() // no longer need exclusive access to the map
				a.rateLimitExceededResponse(w, r, time.Second)
				return
			}

			delay := reservation.Delay()
			if delay > 0 {
				reservation.Cancel()
				mu.Unlock() // no longer need exclusive access to the map
				a.rateLimitExceededResponse(w, r, delay)
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
		mw := newMetricsResponseWriter(w)
		next.ServeHTTP(mw, r)
		durationMs := float64(time.Since(start).Microseconds()) / 1000.0
		a.logger.Info("request received",
			"method", r.Method,
			"path", r.URL.Path,
			"remote_addr", r.RemoteAddr,
			"status", mw.statusCode,
			"duration_ms", durationMs,
		)
	})
}

// gzipResponseWriter wraps an http.ResponseWriter to compress output with gzip
type gzipResponseWriter struct {
	wrapped       http.ResponseWriter
	gzipWriter    *gzip.Writer
	headerWritten bool
}

// newGzipResponseWriter creates a new gzipResponseWriter
func newGzipResponseWriter(w http.ResponseWriter) *gzipResponseWriter {
	gzipWriter := gzip.NewWriter(w)
	return &gzipResponseWriter{
		wrapped:    w,
		gzipWriter: gzipWriter,
	}
}

// Header returns the header map
func (w *gzipResponseWriter) Header() http.Header {
	return w.wrapped.Header()
}

// WriteHeader writes the status code and sets Content-Encoding: gzip
func (w *gzipResponseWriter) WriteHeader(statusCode int) {
	if w.headerWritten {
		return
	}
	w.wrapped.Header().Set("Content-Encoding", "gzip")
	w.wrapped.Header().Del("Content-Length")
	w.wrapped.WriteHeader(statusCode)
	w.headerWritten = true
}

// Write compresses the data using gzip
func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	if !w.headerWritten {
		w.WriteHeader(http.StatusOK)
	}
	return w.gzipWriter.Write(b)
}

// flush flushes the gzip writer to ensure all data is written
func (w *gzipResponseWriter) flush() error {
	return w.gzipWriter.Close()
}

// gzip middleware compresses response bodies with gzip when the client supports it
func (a *applicationDependencies) gzip(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// check if the client accepts gzip encoding
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		// wrap the response writer with gzip compression
		gzipWriter := newGzipResponseWriter(w)
		defer gzipWriter.flush()

		next.ServeHTTP(gzipWriter, r)
	})
}
