package common

import (
	"net"
	"os"
)

// GetPublicIP get current server's public IP address
func GetPublicIP() (string, error) {
	// First try to get from environment variable
	if publicIP := os.Getenv("PUBLIC_IP"); publicIP != "" {
		return publicIP, nil
	}

	// If not in environment variable, try to get local IP
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				// Exclude private IP addresses
				if !isPrivateIP(ipnet.IP) {
					return ipnet.IP.String(), nil
				}
			}
		}
	}

	// If no public IP found, return first non-loopback IPv4 address
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}

	return "", nil
}

// isPrivateIP check if IP is a private IP
func isPrivateIP(ip net.IP) bool {
	// Private IP address ranges
	privateIPBlocks := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
	}

	for _, block := range privateIPBlocks {
		_, cidr, _ := net.ParseCIDR(block)
		if cidr.Contains(ip) {
			return true
		}
	}

	return false
}
