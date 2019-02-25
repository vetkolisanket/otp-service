package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	server := &http.Server{
		Addr:              ":1234",
		Handler:           nil,
		ReadHeaderTimeout: 2 * time.Second,
		WriteTimeout:      1 * time.Minute,
	}

	log.Fatal(server.ListenAndServe())
}

//NewHTTP handles routing of api requests for otp-service
func NewHTTP() http.HandlerFunc {
	mux := http.NewServeMux()

	mux.HandleFunc("/otp-service/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{"success":true}`)
	})
	return mux.ServeHTTP(w, r)
}
