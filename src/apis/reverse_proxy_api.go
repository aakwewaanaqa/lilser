package apis

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func UseReverseProxy(revProxy string, port int) {
	// Define the target URL where requests will be proxied to
	targetURL, err := url.Parse(revProxy)
	if err != nil {
		log.Fatalf("Failed to parse target URL: %v", err)
	}

	// Create a new ReverseProxy
	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	// Define the handler for the proxy server
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Log the incoming request
		log.Printf("Proxying request from %s to %s%s", r.RemoteAddr, targetURL.Host, r.URL.Path)
		// Serve the request through the reverse proxy
		proxy.ServeHTTP(w, r)
	})

	// Start the proxy server on a different port (e.g., 8000)
	// Requests to localhost:8000 will be forwarded to localhost:8080
	proxyPort := fmt.Sprintf(":%d", port)
	log.Printf("Proxy server starting on %s, forwarding to %s", proxyPort, targetURL.String())
}
