package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
)

// Backend connects to a backend (https)
// and forward requests from frontend
type Backend interface {
	IsPrimary() (bool, error)
	IsReadWrite() (bool, error)
	IsReplica() (bool, error)
	IsReadOnly() (bool, error)
}

// NewBackend creates a backend from a driver, a connection string
// and an optional interval for caching values
func NewBackend(config BackendConfig, cache Cache) Backend {
	if config.Host == "" {
		config.Host = "localhost"
	}
	if config.Port == 0 {
		config.Port = 80
	}
	if config.Scheme == "" {
		config.Scheme = "http"
	}

	b := &HTTPBackend{
		host:   config.Host,
		port:   config.Port,
		scheme: config.Scheme,
		cache:  cache,
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: config.Insecure},
	}
	b.client = &http.Client{Transport: tr}
	return b
}

// HTTPBackend will request a backend with HTTP(s) protocol
type HTTPBackend struct {
	cache  Cache
	host   string
	port   int
	scheme string
	client *http.Client
}

func (b HTTPBackend) baseURL() string {
	return fmt.Sprintf("%s://%s:%d", b.scheme, b.host, b.port)
}

// request search into the cache to find the given key
// then eventually creates a HTTP/HTTPS request on backend and
// cache response for further requests
func (b HTTPBackend) request(key string) (bool, error) {
	state, err := b.cache.Get(key)
	if err != nil {
		Warning("could not get key %s from cache: %v", key, err)
		return false, err
	}

	if state == nil {
		url := b.baseURL() + "/" + key

		Debug("GET %s", url)
		response, err := b.client.Get(url)
		if err != nil {
			Warning("could not request remote backend: %v", err)
			return false, err
		}
		defer response.Body.Close()

		if response.StatusCode == http.StatusOK {
			state = true
		} else {
			state = false
		}

		err = b.cache.Set(key, state)
		if err != nil {
			Warning("could not save %s key to cache: %v", key, err)
			return false, err
		}
	}

	return state.(bool), nil
}

// IsPrimary will call /primary route on patroni API
func (b HTTPBackend) IsPrimary() (bool, error) {
	return b.request("primary")
}

// IsReadWrite will call /read-write route on patroni API
func (b HTTPBackend) IsReadWrite() (bool, error) {
	return b.request("read-write")
}

// IsReplica will call /replica route on patroni API
func (b HTTPBackend) IsReplica() (bool, error) {
	return b.request("replica")
}

// IsReadOnly will call /read-only route on patroni API
func (b HTTPBackend) IsReadOnly() (bool, error) {
	return b.request("read-only")
}
