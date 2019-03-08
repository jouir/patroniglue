package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

// Frontend exposes statuses over HTTP(S)
type Frontend struct {
	backend  Backend
	host     string
	port     int
	certfile string
	keyfile  string
}

var backend Backend
var logFormat string

// Start creates an HTTP server and listen
func (f *Frontend) Start() error {
	Debug("creating router")
	r := mux.NewRouter()
	r.Use(loggingMiddleware)
	r.Use(headersMiddleware)

	Debug("registering routes")
	r.HandleFunc("/health", HealthHandler).Methods("GET")
	r.HandleFunc("/master", PrimaryHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/replica", ReplicaHandler).Methods("GET", "OPTIONS")

	Info("listening on %s", f)
	var err error
	if f.certfile != "" && f.keyfile != "" {
		err = http.ListenAndServeTLS(f.String(), f.certfile, f.keyfile, r)
	} else {
		err = http.ListenAndServe(f.String(), r)
	}

	if err != nil {
		return err
	}

	return nil
}

func (f *Frontend) String() string {
	return fmt.Sprintf("%s:%d", f.host, f.port)
}

// NewFrontend creates a Frontend
func NewFrontend(config FrontendConfig, b Backend) (*Frontend, error) {
	backend = b
	logFormat = config.LogFormat
	return &Frontend{
		host:     config.Host,
		port:     config.Port,
		certfile: config.Certfile,
		keyfile:  config.Keyfile,
	}, nil
}

// Log requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Info(formatRequest(r, logFormat))
		next.ServeHTTP(w, r)
	})
}

// formatRequest replaces request placeholders for logging purpose
func formatRequest(r *http.Request, format string) string {
	if format == "" {
		format = "%a - %m %U"
	}
	definitions := map[string]string{
		"%a": r.RemoteAddr,
		"%m": r.Method,
		"%U": r.RequestURI,
	}
	output := format

	for placeholder, value := range definitions {
		output = strings.Replace(output, placeholder, value, -1)
	}

	return output
}

// Add headers
func headersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

// HealthHandler returns frontend health status
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, `{"healthy": true}`)
}

// PrimaryHandler exposes primary status
func PrimaryHandler(w http.ResponseWriter, r *http.Request) {
	var message string
	var status int
	primary, err := backend.IsPrimary()
	if err != nil {
		message = fmt.Sprintf("{\"error\":\"%v\"}", err)
		status = http.StatusServiceUnavailable
	}
	message = fmt.Sprintf("{\"primary\":%t}", primary)
	status = http.StatusServiceUnavailable
	if primary {
		status = http.StatusOK
	}
	w.WriteHeader(status)
	io.WriteString(w, message)
}

// ReplicaHandler exposes replica status
func ReplicaHandler(w http.ResponseWriter, r *http.Request) {
	var message string
	var status int
	replica, err := backend.IsReplica()
	if err != nil {
		message = fmt.Sprintf("{\"error\":\"%v\"}", err)
		status = http.StatusServiceUnavailable
	}
	message = fmt.Sprintf("{\"replica\":%t}", replica)
	status = http.StatusServiceUnavailable
	if replica {
		status = http.StatusOK
	}
	w.WriteHeader(status)
	io.WriteString(w, message)
}
