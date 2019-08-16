package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

// Frontend exposes statuses over HTTP(S)
type Frontend struct {
	backend       Backend
	host          string
	port          int
	certfile      string
	keyfile       string
	tlsMinVersion string
	tlsCiphers    []string
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
	r.HandleFunc("/primary", PrimaryHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/read-write", ReadWriteHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/replica", ReplicaHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/read-only", ReadOnlyHandler).Methods("GET", "OPTIONS")

	Info("listening on %s", f)
	var err error
	server := &http.Server{
		Addr:    f.String(),
		Handler: r,
	}
	if f.certfile != "" && f.keyfile != "" {
		config := &tls.Config{}
		if f.tlsMinVersion != "" {
			config.MinVersion, err = parseTLSVersion(f.tlsMinVersion)
			if err != nil {
				return err
			}
		}
		if len(f.tlsCiphers) > 0 {
			ciphers, err := parseCiphersSuite(f.tlsCiphers)
			if err != nil {
				return err
			}
			config.CipherSuites = ciphers
		}

		server.TLSConfig = config
		err = server.ListenAndServeTLS(f.certfile, f.keyfile)
	} else {
		err = server.ListenAndServe()
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
		host:          config.Host,
		port:          config.Port,
		certfile:      config.Certfile,
		keyfile:       config.Keyfile,
		tlsMinVersion: config.TLSMinVersion,
		tlsCiphers:    config.TLSCiphers,
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

// ReadWriteHandler exposes read-write status
func ReadWriteHandler(w http.ResponseWriter, r *http.Request) {
	var message string
	var status int
	readWrite, err := backend.IsReadWrite()
	if err != nil {
		message = fmt.Sprintf("{\"error\":\"%v\"}", err)
		status = http.StatusServiceUnavailable
	}
	message = fmt.Sprintf("{\"read-write\":%t}", readWrite)
	status = http.StatusServiceUnavailable
	if readWrite {
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

// ReadOnlyHandler exposes read-only status
func ReadOnlyHandler(w http.ResponseWriter, r *http.Request) {
	var message string
	var status int
	readOnly, err := backend.IsReadOnly()
	if err != nil {
		message = fmt.Sprintf("{\"error\":\"%v\"}", err)
		status = http.StatusServiceUnavailable
	}
	message = fmt.Sprintf("{\"read-only\":%t}", readOnly)
	status = http.StatusServiceUnavailable
	if readOnly {
		status = http.StatusOK
	}
	w.WriteHeader(status)
	io.WriteString(w, message)
}

// Store TLS ciphers map from string to constant
// See full list at https://golang.org/pkg/crypto/tls/#pkg-constants
var tlsCiphers = map[string]uint16{
	"TLS_RSA_WITH_RC4_128_SHA":                tls.TLS_RSA_WITH_RC4_128_SHA,
	"TLS_RSA_WITH_3DES_EDE_CBC_SHA":           tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA,
	"TLS_RSA_WITH_AES_128_CBC_SHA":            tls.TLS_RSA_WITH_AES_128_CBC_SHA,
	"TLS_RSA_WITH_AES_256_CBC_SHA":            tls.TLS_RSA_WITH_AES_256_CBC_SHA,
	"TLS_RSA_WITH_AES_128_CBC_SHA256":         tls.TLS_RSA_WITH_AES_128_CBC_SHA256,
	"TLS_RSA_WITH_AES_128_GCM_SHA256":         tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
	"TLS_RSA_WITH_AES_256_GCM_SHA384":         tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
	"TLS_ECDHE_ECDSA_WITH_RC4_128_SHA":        tls.TLS_ECDHE_ECDSA_WITH_RC4_128_SHA,
	"TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA":    tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
	"TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA":    tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
	"TLS_ECDHE_RSA_WITH_RC4_128_SHA":          tls.TLS_ECDHE_RSA_WITH_RC4_128_SHA,
	"TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA":     tls.TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA,
	"TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA":      tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
	"TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA":      tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
	"TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256": tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256,
	"TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256":   tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256,
	"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256":   tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
	"TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256": tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
	"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384":   tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
	"TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384": tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
	"TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305":    tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
	"TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305":  tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
	"TLS_AES_128_GCM_SHA256":                  tls.TLS_AES_128_GCM_SHA256,
	"TLS_AES_256_GCM_SHA384":                  tls.TLS_AES_256_GCM_SHA384,
	"TLS_CHACHA20_POLY1305_SHA256":            tls.TLS_CHACHA20_POLY1305_SHA256,
	"TLS_FALLBACK_SCSV":                       tls.TLS_FALLBACK_SCSV,
}

// Convert a list of ciphers from string to TLS constants
func parseCiphersSuite(strings []string) (ciphers []uint16, err error) {
	for _, s := range strings {
		if cipher, ok := tlsCiphers[s]; ok {
			ciphers = append(ciphers, cipher)
		} else {
			return nil, fmt.Errorf("unknown cipher detected: %s", s)
		}
	}
	return ciphers, nil
}

// Store TLS versions map from string to constant
// See full list at https://golang.org/pkg/crypto/tls/#pkg-constants
var tlsVersions = map[string]uint16{
	"SSLv3.0": tls.VersionSSL30,
	"TLSv1.0": tls.VersionTLS10,
	"TLSv1.1": tls.VersionTLS11,
	"TLSv1.2": tls.VersionTLS12,
	"TLSv1.3": tls.VersionTLS13,
}

// Convert a list of ciphers from string to TLS constants
func parseTLSVersion(s string) (uint16, error) {
	if version, ok := tlsVersions[s]; ok {
		return version, nil
	}
	return 0, fmt.Errorf("unknown TLS version: %s", s)
}
