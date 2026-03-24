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
	<title>Medical System Preflight CORS Example</title>
</head>
<body>
	<h1>Medical Appointment Scheduling System</h1>
	<p>
		This page demonstrates a preflight CORS request from a frontend running on a different origin.
		It attempts to create a patient by sending a POST request with JSON to the API.
		Because the request uses the Content-Type application/json header, the browser will first send
		an OPTIONS preflight request to ask the server whether the real POST request is allowed.
	</p>

	<h2>Result</h2>
	<div id="output">Loading...</div>

	<script>
	document.addEventListener('DOMContentLoaded', function() {
		fetch("http://localhost:4000/v1/patients", {
			method: "POST",
			headers: {
				"Content-Type": "application/json"
			},
			body: JSON.stringify({
				first_name: "Preflight",
				last_name: "Test",
				date_of_birth: "2003-05-14",
				gender: "Female",
				phone: "501-600-1234",
				email: "test@example.com"
			})
		})
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

	log.Printf("Starting preflight CORS example server on %s", *addr)

	err := http.ListenAndServe(*addr, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(html))
	}))
	log.Fatal(err)
}
