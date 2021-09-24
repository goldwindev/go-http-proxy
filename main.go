package main

import (
	"crypto/tls"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/alexcesaro/log"
	"github.com/alexcesaro/log/stdlog"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"

	"github.com/koding/websocketproxy"
)

// Get env var or default
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// Log the typeform payload and redirect url
func logRequestPayload(proxyUrl string, req *http.Request) {
	Logger.Infof("Origin: %s | Path: %s", proxyUrl, req.URL.String())
}

// Log the env variables required for a reverse proxy
func logSetup() {
	proto := "HTTP"
	if ShouldUseTLS {
		proto = "HTTPS"
	}

	Logger.Infof("%s Server Listening on: %s", proto, ListenAddress)
	Logger.Infof("Redirecting to url: %s", ProxyURL)
}

// Serve a reverse proxy for a given url
func serveReverseProxy(target string, res http.ResponseWriter, req *http.Request) {
	url, err := url.Parse(target)
	if err != nil {
		Logger.Error(err)
		return
	}

	// create the reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(url)

	proxy.Director = func(r *http.Request) {
		b, err := httputil.DumpRequest(r, true)
		if err != nil {
			Logger.Error(err)
			return
		}
		Logger.Debug("==============REQUEST_START=============")
		Logger.Debug("\n" + string(b))
		Logger.Debug("===============REQUEST_END==============")
	}

	proxy.ModifyResponse = func(r *http.Response) error {
		b, err := httputil.DumpResponse(r, true)
		if err != nil {
			Logger.Error(err)
			return err
		}
		Logger.Debug("=============RESPONSE_START=============")
		Logger.Debug("\n" + string(b))
		Logger.Debug("==============RESPONSE_END==============")

		return nil
	}

	isWebsocket := req.Header.Get("Connection") == "Upgrade"
	var wsProxy *websocketproxy.WebsocketProxy
	if isWebsocket {
		url.Scheme = strings.Replace(url.Scheme, "http", "ws", 1) // http(s) -> ws(s)
		wsProxy = websocketproxy.NewProxy(url)
		wsProxy.Upgrader = &websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		}
	}

	// Update the headers to allow for SSL redirection
	req.URL.Host = url.Host
	req.URL.Scheme = url.Scheme
	forwardedHost := req.Header.Get("Host")
	if forwardedHost != "" {
		req.Header.Set("X-Forwarded-Host", forwardedHost)
	}
	req.Host = url.Host

	if isWebsocket {
		wsProxy.ServeHTTP(res, req)
	} else {
		proxy.ServeHTTP(res, req)
	}
}

// Given a request send it to the appropriate url
func handleRequestAndRedirect(res http.ResponseWriter, req *http.Request) {
	url := ProxyURL

	logRequestPayload(url, req)

	serveReverseProxy(url, res, req)
}

func createServer() *http.Server {
	s := &http.Server{
		Addr:    ListenAddress,
		Handler: nil, // use `http.DefaultServeMux`
	}

	if ShouldUseTLS {
		// generate a `Certificate` struct
		cert, _ := tls.LoadX509KeyPair(SSLCert, SSLKey)
		Logger.Infof("Loaded X509 KeyPair (%s / %s)", SSLCert, SSLKey)

		// create a custom server with `TLSConfig`
		s.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
		}
	}

	return s
}

var (
	ListenAddress string
	ProxyURL      string
	ShouldUseTLS  bool
	SSLCert       string
	SSLKey        string
	LogLevel      int
	Logger        log.Logger
)

func init() {
	Logger = stdlog.GetFromFlags()

	// load env vars from .env file if exists, else rely on system vars or defaults
	err := godotenv.Load()
	if err != nil {
		Logger.Warning("Failed to load .env file")
	}

	ListenAddress = ":" + getEnv("PORT", "1337")
	ProxyURL = getEnv("PROXY_URL", "http://127.0.0.1:31337")
	ShouldUseTLS = getEnv("USE_TLS", "") == "true"
	SSLCert = getEnv("SSL_CERT", "")
	SSLKey = getEnv("SSL_KEY", "")
}

func main() {
	// Configure server
	http.HandleFunc("/", handleRequestAndRedirect)
	s := createServer()

	// Log setup values
	logSetup()

	// Start server
	if ShouldUseTLS {
		Logger.Error(s.ListenAndServeTLS("", ""))
	} else {
		Logger.Error(s.ListenAndServe())
	}
}
