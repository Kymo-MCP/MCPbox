package utils

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

// GetHostIPs gets all network interface IP addresses excluding loopback addresses
// and fetches public IP address from network service
func GetHostIPs() ([]string, error) {
	var ips []string

	// Get local network interface IP addresses
	localIPs, err := getLocalIPs()
	if err != nil {
		return nil, fmt.Errorf("failed to get local IPs: %w", err)
	}
	ips = append(ips, localIPs...)

	// Get public IP address
	publicIP, err := getPublicIPFromService()
	if err != nil {
		// Log error but don't fail the entire function
		fmt.Printf("Warning: failed to get public IP: %v\n", err)
	} else if publicIP != "" {
		ips = append(ips, publicIP)
	}

	return ips, nil
}

// getLocalIPs gets all local network interface IP addresses excluding loopback
func getLocalIPs() ([]string, error) {
	var ips []string

	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to get network interfaces: %w", err)
	}

	for _, iface := range interfaces {
		// Skip loopback and down interfaces
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			// Skip loopback IPs and non-IPv4 addresses
			if ip == nil || ip.IsLoopback() {
				continue
			}

			// Only include IPv4 addresses
			if ip.To4() != nil {
				ips = append(ips, ip.String())
			}
		}
	}

	return ips, nil
}

// getPublicIPFromService fetches public IP from external service
func getPublicIPFromService() (string, error) {
	// List of public IP services to try
	services := []string{
		"https://api.ipify.org",
		"https://ipinfo.io/ip",
		"https://icanhazip.com",
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	for _, service := range services {
		resp, err := client.Get(service)
		if err != nil {
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			continue
		}

		if resp.StatusCode == http.StatusOK {
			ip := strings.TrimSpace(string(body))
			// Validate IP format
			if net.ParseIP(ip) != nil {
				return ip, nil
			}
		}
	}

	return "", fmt.Errorf("failed to get public IP from all services")
}
