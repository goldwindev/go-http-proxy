package main

import (
	"crypto/tls"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/joho/godotenv"
)

// Get env var or default
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// Get the port to listen on
func getListenAddress() string {
	return ":" + getEnv("PORT", "1337")
}

// Get the URL to redirect to
func getProxyURL() string {
	return getEnv("PROXY_URL", "127.0.0.1:31337")
}

// Get the URL to redirect to
func shouldUseTLS() bool {
	return getEnv("USE_TLS", "") != ""
}

// Log the typeform payload and redirect url
func logRequestPayload(proxyUrl string, req *http.Request) {
	log.Printf("Origin: %s | Path: %s\n", proxyUrl, req.URL.String())
}

// Log the env variables required for a reverse proxy
func logSetup() {
	log.Printf("Server Listening on: %s\n", getListenAddress())
	log.Printf("Redirecting to url: %s\n", getProxyURL())
}

// Serve a reverse proxy for a given url
func serveReverseProxy(target string, res http.ResponseWriter, req *http.Request) {
	url, _ := url.Parse(target)

	// create the reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(url)

	// Update the headers to allow for SSL redirection
	req.URL.Host = url.Host
	req.URL.Scheme = url.Scheme
	req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
	req.Host = url.Host

	proxy.ServeHTTP(res, req)
}

// Given a request send it to the appropriate url
func handleRequestAndRedirect(res http.ResponseWriter, req *http.Request) {
	url := getProxyURL()

	logRequestPayload(url, req)

	serveReverseProxy(url, res, req)
}

func createServer() *http.Server {
	s := &http.Server{
		Addr:    getListenAddress(),
		Handler: nil, // use `http.DefaultServeMux`
	}

	if shouldUseTLS() {
		// generate a `Certificate` struct
		cert, _ := tls.LoadX509KeyPair(getEnv("SSL_CERT", ""), getEnv("SSL_KEY", ""))

		// create a custom server with `TLSConfig`
		s.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
		}
	}

	return s
}

func main() {
	// load env vars from .env file if exists, else rely on system vars or defaults
	godotenv.Load()

	// Log setup values
	logSetup()

	// start server
	http.HandleFunc("/", handleRequestAndRedirect)

	s := createServer()
	if shouldUseTLS() {
		log.Fatal(s.ListenAndServeTLS("", ""))
	} else {
		log.Fatal(s.ListenAndServe())
	}
}
