package proxy

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type ProxyServer struct {
	runtimeEndpoint       string
	port                  int
	server                *http.Server
	EventsChannel         chan *http.Response
	LambdaResponseChannel chan *http.Request
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
		EventsChannel:         make(chan *http.Response),
		LambdaResponseChannel: make(chan *http.Request),
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
		&ps.EventsChannel,
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
		&ps.LambdaResponseChannel,
		nil,
	)
	r.Post("/2018-06-01/runtime/invocation/{requestId}/response", responseHandler)

	ps.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", ps.port),
		Handler: r,
	}

	return ps, nil
}

func (p *ProxyServer) ListenAndServe() error {
	return p.server.ListenAndServe()
}
