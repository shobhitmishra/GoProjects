package main

import (
	b64 "encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

const fileName = "urlmapping.json"

type urlMap struct {
	Mapping map[string]string `json:"urlMap"`
}

// This method is used in two places
// 1. When we get a post request for a  new url shortening.
// We verify if the url is valid or not before creating and writing a shortened url.
// 2. When we get a url redirection request, the url param can be for one of the already shortened urls
// or for a new url (such as https://www.wizardingworld.com). If it is non stored shortened url,
// we verify that it is a valid and reachable url.
func verifyUrl(url string) error {
	client := http.Client{
		Timeout: 1 * time.Second,
	}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return errors.New("url returned a non 200 status")
	}
	return nil
}

// Reads the content of the file <fileName> and deserializes it in a urlmap
// File should be present else we'll get an error message. The caller will get a well-formed struct though.
func readUrlMap() urlMap {
	umap := urlMap{Mapping: map[string]string{}}
	if byteVal, fileReadErr := os.ReadFile(fileName); fileReadErr != nil {
		log.Printf("can't open the file %v", fileReadErr)
	} else {
		if len(byteVal) == 0 {
			return umap
		}
		if err := json.Unmarshal(byteVal, &umap); err != nil {
			log.Fatalf("can't decode the json file %v", err)
		}
	}
	return umap
}

// Writes the content of the urlMapping into file <fileName>.
// The file should already exist for it to work.
// TODO: Create file if it doesn't exist.
func writeUrlMap(urlMapping urlMap) {
	if mappingJson, err := json.Marshal(urlMapping); err != nil {
		log.Fatalf("Couldn't encode the urlMapping to json %v\n", err)
	} else {
		if err = os.WriteFile(fileName, mappingJson, 0644); err != nil {
			log.Fatalf("Can't write the json file %v\n", err)
		}
	}
}

// Used by handleNewUrl. Read the request body for a url,
// Verify the url is valid and reachable and encode it using base64 encoding scheme.
func shortenUrl(url []byte) string {
	sEnc := b64.StdEncoding.EncodeToString(url)
	return sEnc[:10]
}

// Handles the new shortening request. i.e. take https://www.google.com
// encode it and write back the new mapping to file.
// We verify that url is valid and reachable before we write it to the file.
// We return a bad request and error if the url is invalid.
// A valid url is reachable. It has to have https:// as prefix.
// Example of good urls are https://www.google.com, https://www.apple.com
// Invoked on POST request.
// Example curl request for this endpoint:
/* curl -X "POST" "http://localhost:8080/api/v1/new" \
-H 'Content-Type: text/plain; charset=utf-8' \
-d "https://www.wikipedia.org" */

func handleNewUrl(w http.ResponseWriter, r *http.Request) {
	// Read the request body
	newUrl, err := io.ReadAll(r.Body)
	if err != nil {
		log.Fatalf("Couldn't read request body")
		return
	}
	if err := verifyUrl(string(newUrl)); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "url is either invalid or not reachable \n Error: %s", err.Error())
		return
	}
	shortenedUrl := shortenUrl(newUrl)
	urlMapping := readUrlMap()
	urlMapping.Mapping[shortenedUrl] = string(newUrl)
	writeUrlMap(urlMapping)
}

// Handles the redirection request.
// If the requested url is one of the stored mappings then it is redirected to the stored url.
// e.g. if one of the stored mapping is "aHR0cHM6Ly":"https://www.wikipedia.org", then this request
// will redirect us to https://www.wikipedia.org.
// If the requested url is not a stored shortened url then we redirect to the requested url.
// e.g. for a shortened url aHR0cHM6Ly, example valid curl is http://localhost:8080/api/v1/url?url=aHR0cHM6Lyyy"
// For a non-shortened but valid url, example curl is
// http://localhost:8080/api/v1/url?url=https%3A%2F%2Fwww.wikipedia.org
// If the requested url is neither a shortened url nor a valid url then we get an error and a bad request.
// NOTE: http.Redirect redirects to a relative url unless the destination url has https:// as prefix.
func redirectUrl(w http.ResponseWriter, r *http.Request) {
	// read url from the query param
	queryParams := r.URL.Query()
	requestedUrl := queryParams["url"][0]
	log.Printf("Got parameter url:%s \n", requestedUrl)
	// read it from the file
	urlMapping := readUrlMap()
	if redirect, ok := urlMapping.Mapping[requestedUrl]; ok {
		http.Redirect(w, r, redirect, http.StatusFound)
	} else {
		// verify requestedUrl
		if err := verifyUrl(requestedUrl); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "url is either invalid or not reachable \n Error: %s", err.Error())
			return
		}
		http.Redirect(w, r, requestedUrl, 302)
	}
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/api/v1/new", handleNewUrl).Methods("POST")
	r.HandleFunc("/api/v1/url", redirectUrl).Methods("GET")

	log.Fatal(http.ListenAndServe(":8080", r))
}
