package utils

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// PortProbeOptions port probe options
type PortProbeOptions struct {
	Host    string        // host address
	Port    int           // port number
	Timeout time.Duration // timeout
}

// PortProbeResult port probe result
type PortProbeResult struct {
	Success bool          // probe success
	Error   string        // error message
	Latency time.Duration // connection latency
}

// HTTPProbeOptions HTTP probe options
type HTTPProbeOptions struct {
	URL     string        // probe URL
	Timeout time.Duration // timeout
	Method  string        // HTTP method, defaults to HEAD
}

// HTTPProbeResult HTTP probe result
type HTTPProbeResult struct {
	Success    bool          // probe success
	StatusCode int           // HTTP status code
	Error      string        // error message
	Latency    time.Duration // response latency
}

// ProbeHTTP204 probe HTTP service availability (expecting 204 status code)
// uses HEAD method to reduce network overhead
func ProbeHTTP204(ctx context.Context, options HTTPProbeOptions) *HTTPProbeResult {
	return ProbeHTTP(ctx, options, 204)
}

// ProbeHTTPHealth probe HTTP service health (expecting 2xx status codes)
func ProbeHTTPHealth(ctx context.Context, options HTTPProbeOptions) *HTTPProbeResult {
	return ProbeHTTP(ctx, options, 0) // 0 means accept any 2xx status code
}

// ProbeHTTP generic HTTP probe function
// when expectedStatus is 0, accepts any 2xx status code; otherwise must match specified status code
func ProbeHTTP(ctx context.Context, options HTTPProbeOptions, expectedStatus int) *HTTPProbeResult {
	start := time.Now()
	result := &HTTPProbeResult{
		Success: false,
	}

	// set defaults
	if options.Timeout == 0 {
		options.Timeout = 5 * time.Second
	}
	if options.Method == "" {
		options.Method = "HEAD" // default to HEAD to reduce network overhead
	}

	// validate URL
	if options.URL == "" {
		result.Error = "URL cannot be empty"
		return result
	}

	// create HTTP client
	client := &http.Client{
		Timeout: options.Timeout,
	}

	// create request
	req, err := http.NewRequestWithContext(ctx, options.Method, options.URL, nil)
	if err != nil {
		result.Error = fmt.Sprintf("failed to create request: %v", err)
		return result
	}

	// set User-Agent
	req.Header.Set("User-Agent", "qm-mcp-server-health-checker/1.0")

	// perform request
	resp, err := client.Do(req)
	if err != nil {
		result.Error = fmt.Sprintf("request failed: %v", err)
		result.Latency = time.Since(start)
		return result
	}
	defer resp.Body.Close()

	// record response time
	result.Latency = time.Since(start)
	result.StatusCode = resp.StatusCode

	// check status code
	if expectedStatus == 0 {
		// accept any 2xx status code
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			result.Success = true
		} else {
			result.Error = fmt.Sprintf("expected 2xx status code, got: %d", resp.StatusCode)
		}
	} else {
		// must match specified status code
		if resp.StatusCode == expectedStatus {
			result.Success = true
		} else {
			result.Error = fmt.Sprintf("expected status code %d, got: %d", expectedStatus, resp.StatusCode)
		}
	}

	return result
}

// ProbePortFromURL extract address from URL and perform port probe
func ProbePortFromURL(ctx context.Context, urlStr string, timeout time.Duration) *PortProbeResult {
	result := &PortProbeResult{
		Success: false,
	}

	// parse URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		result.Error = fmt.Sprintf("failed to parse URL: %v", err)
		return result
	}

	// extract host and port
	host := parsedURL.Hostname()
	if host == "" {
		result.Error = "no valid hostname found in URL"
		return result
	}

	// get port
	portStr := parsedURL.Port()
	var port int
	if portStr == "" {
		// set default port based on scheme
		switch parsedURL.Scheme {
		case "http":
			port = 80
		case "https":
			port = 443
		default:
			result.Error = fmt.Sprintf("cannot determine port, scheme: %s", parsedURL.Scheme)
			return result
		}
	} else {
		port, err = strconv.Atoi(portStr)
		if err != nil {
			result.Error = fmt.Sprintf("invalid port format: %v", err)
			return result
		}
	}

	// perform port probe
	return ProbePort(ctx, PortProbeOptions{
		Host:    host,
		Port:    port,
		Timeout: timeout,
	})
}

// ProbePort probe connectivity of specified host and port
func ProbePort(ctx context.Context, options PortProbeOptions) *PortProbeResult {
	start := time.Now()
	result := &PortProbeResult{
		Success: false,
	}

	// set default timeout
	if options.Timeout == 0 {
		options.Timeout = 5 * time.Second
	}

	// validate parameters
	if options.Host == "" {
		result.Error = "host address cannot be empty"
		return result
	}
	if options.Port <= 0 || options.Port > 65535 {
		result.Error = fmt.Sprintf("invalid port number: %d", options.Port)
		return result
	}

	// build address
	address := fmt.Sprintf("%s:%d", options.Host, options.Port)

	// create dialer with timeout
	dialer := &net.Dialer{
		Timeout: options.Timeout,
	}

	// attempt connection
	conn, err := dialer.DialContext(ctx, "tcp", address)
	result.Latency = time.Since(start)

	if err != nil {
		result.Error = fmt.Sprintf("connection failed: %v", err)
		return result
	}

	// connection successful, close immediately
	conn.Close()
	result.Success = true
	return result
}

// IsHTTPServiceAvailable convenience function to check if HTTP service is available
func IsHTTPServiceAvailable(ctx context.Context, url string, timeout time.Duration) bool {
	result := ProbeHTTPHealth(ctx, HTTPProbeOptions{
		URL:     url,
		Timeout: timeout,
	})
	return result.Success
}

// IsPortAvailable convenience function to check if port is available
func IsPortAvailable(ctx context.Context, host string, port int, timeout time.Duration) bool {
	result := ProbePort(ctx, PortProbeOptions{
		Host:    host,
		Port:    port,
		Timeout: timeout,
	})
	return result.Success
}

// IsPortAvailableFromURL convenience function to extract address from URL and check if port is available
func IsPortAvailableFromURL(ctx context.Context, urlStr string, timeout time.Duration) bool {
	result := ProbePortFromURL(ctx, urlStr, timeout)
	return result.Success
}

// IsHTTP204Available convenience function to check if HTTP service returns 204 status code
func IsHTTP204Available(ctx context.Context, url string, timeout time.Duration) bool {
	result := ProbeHTTP204(ctx, HTTPProbeOptions{
		URL:     url,
		Timeout: timeout,
	})
	return result.Success
}
