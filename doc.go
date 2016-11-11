/*
Package kms (Killing Me Softly) is a library that aids in graceful shutdown of a process/application.

Example:
	package main

	import (
		"net/http"
		"time"

		"github.com/go-playground/kms"
		"github.com/go-playground/kms/kmsnet/kmshttp"
	)

	func main() {

		// listen for shutdown signal(s) with timeout, non-blocking
		kms.ListenTimeout(false, time.Minute*3)

		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Home"))
		})

		kmshttp.ListenAndServe(":3007", nil)
	}
*/
package kms
