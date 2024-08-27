package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

func main() {

	address := flag.String("l", "127.0.0.1", "listen address")
	port := flag.Int("p", 19988, "listen port")
	flag.Parse()
	log.Printf("Server listening on http://%s:%d\n", *address, *port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%d", *address, *port), &CustomHandler{}))

}

// CustomHandler implements the http.Handler interface
type CustomHandler struct{}

func (h *CustomHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.String(), "/") {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Invalid request"))
		return
	}

	// Extract the URL to proxy to
	proxyURL := r.URL.String()[1:] // Remove the leading '/'
	parsedURL, err := url.Parse(proxyURL)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Invalid URL"))
		return
	}

	// Create the request to the target server
	proxyReq, err := http.NewRequest(r.Method, parsedURL.String(), r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Failed to create request"))
		return
	}

	// Copy the headers from the original request
	proxyReq.Header = r.Header.Clone()

	// Perform the request to the target server
	client := &http.Client{}
	resp, err := client.Do(proxyReq)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte("Failed to reach target server"))
		return
	}
	defer resp.Body.Close()
	//的
	// Copy the headers from the target server's response
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Copy the status code and body
	w.WriteHeader(resp.StatusCode)
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		log.Printf("Failed to write response: %v", err)
	}
}
