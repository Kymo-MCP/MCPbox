package common

import (
	"net"
	"os"
)

// GetPublicIP 获取当前服务器的公网IP地址
func GetPublicIP() (string, error) {
	// 首先尝试从环境变量获取
	if publicIP := os.Getenv("PUBLIC_IP"); publicIP != "" {
		return publicIP, nil
	}

	// 如果环境变量中没有，则尝试获取本地IP
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				// 排除私有IP地址
				if !isPrivateIP(ipnet.IP) {
					return ipnet.IP.String(), nil
				}
			}
		}
	}

	// 如果没有找到公网IP，返回第一个非回环的IPv4地址
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}

	return "", nil
}

// isPrivateIP 检查IP是否为私有IP
func isPrivateIP(ip net.IP) bool {
	// 私有IP地址范围
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
