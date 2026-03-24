package main

import (
	"flag"
	"log"
	"net/http"
)

const html = `
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<title>Medical System Basic CORS Example</title>
</head>
<body>
	<h1>Medical Appointment Scheduling System</h1>
	<p>
		This page demonstrates a basic CORS request from a frontend running on a different origin.
		It sends a simple GET request to the API healthcheck endpoint to confirm that the backend is online.
		Because this is a simple request, the browser does not need to send a preflight OPTIONS request first.
	</p>

	<h2>Result</h2>
	<div id="output">Loading...</div>

	<script>
	document.addEventListener('DOMContentLoaded', function() {
		fetch("http://localhost:4000/v1/healthcheck")
			.then(function(response) {
				return response.text();
			})
			.then(function(text) {
				document.getElementById("output").innerHTML = text;
			})
			.catch(function(err) {
				document.getElementById("output").innerHTML = "Request failed: " + err;
			});
	});
	</script>
</body>
</html>
`

func main() {
	addr := flag.String("addr", ":9000", "Server address")
	flag.Parse()

	log.Printf("starting basic CORS example server on %s", *addr)

	err := http.ListenAndServe(*addr, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(html))
	}))
	log.Fatal(err)
}
