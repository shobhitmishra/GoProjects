package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// In the request middlewares, you first do stuff then handle http request
// i.e. write middleware code and then call next.ServeHTTP(w, r)
func logRequestMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Do stuff here
		fmt.Fprintf(w, "request uri is: %s, at time: %s \n", r.RequestURI, time.Now())
		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})
}

// In the response middlewares, you first handle http request then do your stuff
// i.e. call next.ServeHTTP(w, r) and then write middleware code
func handleResponseMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
		// Do stuff here
		fmt.Fprintf(w, "responding the request at time: %s \n", time.Now())
	})
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "In the home handler function \n")
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/", homeHandler)
	router.Use(logRequestMiddleware, handleResponseMiddleware)
	log.Fatal(http.ListenAndServe(":8080", router))
}
