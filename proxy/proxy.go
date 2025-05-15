package proxy

import (
	"context"
	"encoding/json"
	"firetail-lambda-extension/firetail"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

type ProxyServer struct {
	runtimeEndpoint       string
	port                  int
	server                *http.Server
	eventsChannel         chan *http.Response
	lambdaResponseChannel chan *http.Request
	RecordsChannel        chan firetail.Record
}

func NewProxyServer() (*ProxyServer, error) {
	portStr, portSet := os.LookupEnv("FIRETAIL_LAMBDA_EXTENSION_PORT")
	var port int
	var err error
	if port, err = strconv.Atoi(portStr); err != nil || !portSet {
		port = 9009
	}

	ps := &ProxyServer{
		runtimeEndpoint:       os.Getenv("AWS_LAMBDA_RUNTIME_API"),
		port:                  port,
		eventsChannel:         make(chan *http.Response),
		lambdaResponseChannel: make(chan *http.Request),
		RecordsChannel:        make(chan firetail.Record),
	}

	r := chi.NewRouter()

	handleError := func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(404), 404)
	}
	r.NotFound(handleError)
	r.MethodNotAllowed(handleError)

	initEndpoint, err := url.Parse(
		fmt.Sprintf(
			"http://%s/2018-06-01/runtime/init/error",
			ps.runtimeEndpoint,
		),
	)
	if err != nil {
		return nil, err
	}
	initErrorHandler := getProxyHandler(
		func(r *http.Request) (*url.URL, error) {
			return initEndpoint, nil
		},
		nil,
		nil,
	)
	r.Post("/2018-06-01/runtime/init/error", initErrorHandler)

	invokeErrorHandler := getProxyHandler(
		func(r *http.Request) (*url.URL, error) {
			return url.Parse(
				fmt.Sprintf(
					"http://%s/2018-06-01/runtime/invocation/%s/error",
					ps.runtimeEndpoint,
					chi.URLParam(r, "requestId"),
				),
			)
		},
		nil,
		nil,
	)
	r.Post("/2018-06-01/runtime/invocation/{requestId}/error", invokeErrorHandler)

	nextEndpoint, err := url.Parse(
		fmt.Sprintf(
			"http://%s/2018-06-01/runtime/invocation/next",
			ps.runtimeEndpoint,
		),
	)
	if err != nil {
		return nil, err
	}
	nextHandler := getProxyHandler(
		func(r *http.Request) (*url.URL, error) {
			return nextEndpoint, nil
		},
		nil,
		&ps.eventsChannel,
	)
	r.Get("/2018-06-01/runtime/invocation/next", nextHandler)

	responseHandler := getProxyHandler(
		func(r *http.Request) (*url.URL, error) {
			return url.Parse(
				fmt.Sprintf(
					"http://%s/2018-06-01/runtime/invocation/%s/response",
					ps.runtimeEndpoint,
					chi.URLParam(r, "requestId"),
				),
			)
		},
		&ps.lambdaResponseChannel,
		nil,
	)
	r.Post("/2018-06-01/runtime/invocation/{requestId}/response", responseHandler)

	ps.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", ps.port),
		Handler: r,
	}

	return ps, nil
}

func (p *ProxyServer) recordAssembler() {
	for {
		// Events and lambda responses should come in pairs, event first and response second.
		event, ok := <-p.eventsChannel
		if !ok {
			log.Println("Events channel closed, stopping record assembler.")
			return
		}

		// We can record the time between receiving the event and the response
		// to calculate the execution time of the lambda function.
		eventReceivedAt := time.Now()

		lambdaResponse, ok := <-p.lambdaResponseChannel
		if !ok {
			log.Println("Lambda response channel closed, stopping record assembler.")
			return
		}

		executionTime := time.Since(eventReceivedAt)

		eventBody, err := io.ReadAll(event.Body)
		if err != nil {
			log.Println("Error reading event body:", err)
			continue
		}
		responseBody, err := io.ReadAll(lambdaResponse.Body)
		if err != nil {
			log.Println("Error reading response body:", err)
			continue
		}

		var recordResponse firetail.RecordResponse
		if err := json.Unmarshal(responseBody, &recordResponse); err != nil {
			log.Println("Error unmarshalling response body:", err)
			continue
		}

		p.RecordsChannel <- firetail.Record{
			Event:         eventBody,
			Response:      recordResponse,
			ExecutionTime: executionTime.Seconds(),
		}
	}
}

func (p *ProxyServer) ListenAndServe() error {
	go p.recordAssembler()
	return p.server.ListenAndServe()
}

func (p *ProxyServer) Shutdown(ctx context.Context) error {
	if err := p.server.Shutdown(ctx); err != nil {
		return err
	}
	close(p.eventsChannel)
	close(p.lambdaResponseChannel)
	return nil
}
