package proxy

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

func getProxyHandler(urlMappingFunc func(r *http.Request) (*url.URL, error), requestChannel *chan *http.Request, responseChannel *chan *http.Response) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the target URL from the mapping function
		targetUrl, err := urlMappingFunc(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Set the request URL to the target URL
		r.RequestURI = ""
		r.Host = targetUrl.Host
		r.URL = targetUrl

		// Make a copy of the request body
		var requestBodyCopy strings.Builder
		r.Body = io.NopCloser(io.TeeReader(r.Body, &requestBodyCopy))

		// Do the request
		resp, err := (&http.Client{}).Do(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Send the request to the requestChannel with the copied body if the channel was provided
		if requestChannel != nil {
			log.Println("Captured lambda response", requestBodyCopy.String())
			r.Body = io.NopCloser(strings.NewReader(requestBodyCopy.String()))
			*requestChannel <- r
		}

		// Make a copy of the response body
		var responseBodyCopy strings.Builder
		resp.Body = io.NopCloser(io.TeeReader(resp.Body, &responseBodyCopy))

		// Write the response to the original response writer
		defer resp.Body.Close()
		for key, value := range resp.Header {
			w.Header()[strings.ToLower(key)] = value
		}
		w.WriteHeader(resp.StatusCode)
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(body)

		// Send the response to the responseChannel with the copied body if the channel was provided
		if responseChannel != nil {
			log.Println("Captured event", responseBodyCopy.String())
			resp.Body = io.NopCloser(strings.NewReader(responseBodyCopy.String()))
			*responseChannel <- resp
		}
	}
}
