package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

var (
	adbPath string
)

func runAdb(args ...string) (string, error) {
	cmd := exec.Command(adbPath, args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

type Device struct {
	Name        string
	IPPort      string
	Type        string // "usb" or "wireless"
	IsConnected bool
}

func DiscoverDevices() ([]Device, error) {
	var devices []Device
	seen := make(map[string]bool) // Track by IP to avoid duplicates

	// Discover all connected devices via adb devices
	usbOutput, err := runAdb("devices")
	if err != nil {
		return nil, fmt.Errorf("failed to check devices: %w", err)
	}
	for _, line := range strings.Split(usbOutput, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "List of") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 2 && parts[1] == "device" {
			serial := parts[0]

			// Skip mDNS entries and invalid serials
			if strings.Contains(serial, "._adb-tls") || serial == "" || serial == "0.0.0.0" {
				continue
			}

			deviceType := "usb"
			ip := serial
			if strings.Contains(serial, ":") {
				deviceType = "wireless"
				// Extract just the IP
				if idx := strings.LastIndex(serial, ":"); idx > 0 {
					ip = serial[:idx]
				}
			}

			// Skip invalid IPs
			if ip == "" || ip == "0.0.0.0" {
				continue
			}

			// Skip duplicates by IP
			if seen[ip] {
				continue
			}
			seen[ip] = true

			devices = append(devices, Device{
				Name:        ip,
				IPPort:      serial,
				Type:        deviceType,
				IsConnected: true,
			})
		}
	}

	// Discover wireless devices via mDNS
	wirelessOutput, err := runAdb("mdns", "services")
	if err != nil {
		return nil, fmt.Errorf("failed to check wireless devices: %w", err)
	}

	for _, line := range strings.Split(wirelessOutput, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "List of") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 3 && strings.Contains(parts[len(parts)-1], ":") {
			ipPort := parts[len(parts)-1]

			// Extract IP
			ip := ipPort
			if idx := strings.LastIndex(ipPort, ":"); idx > 0 {
				ip = ipPort[:idx]
			}

			// Skip invalid IPs
			if ip == "" || ip == "0.0.0.0" {
				continue
			}

			// Skip if already in list
			if seen[ip] {
				continue
			}
			seen[ip] = true

			// Try to connect
			connectOutput, _ := runAdb("connect", ipPort)
			isConnected := strings.Contains(connectOutput, "connected") || strings.Contains(connectOutput, "already connected")
			devices = append(devices, Device{
				Name:        ip,
				IPPort:      ipPort,
				Type:        "wireless",
				IsConnected: isConnected,
			})
		}
	}

	return devices, nil
}

type PairResult struct {
	Success     bool
	ConnectAddr string
	Error       string
}

func PairDevice(ip, port, code string) PairResult {
	// Pair with the device
	output, err := runAdb("pair", ip+":"+port, code)
	if err != nil {
		return PairResult{Success: false, Error: err.Error()}
	}

	if !strings.Contains(strings.ToLower(output), "success") {
		return PairResult{Success: false, Error: output}
	}

	// Find the device's actual port via mDNS
	mdnsOutput, _ := runAdb("mdns", "services")
	for _, line := range strings.Split(mdnsOutput, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "List of") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 3 && strings.Contains(parts[len(parts)-1], ":") && strings.Contains(line, ip) {
			return PairResult{
				Success:     true,
				ConnectAddr: parts[len(parts)-1],
			}
		}
	}

	// Fallback to port 5555
	return PairResult{
		Success:     true,
		ConnectAddr: ip + ":5555",
	}
}

type ConnectResult struct {
	Success bool
	Error   string
}

func ConnectDevice(ipPort string) ConnectResult {
	output, err := runAdb("connect", ipPort)
	if err != nil {
		return ConnectResult{Success: false, Error: err.Error()}
	}

	if strings.Contains(output, "connected") || strings.Contains(output, "already connected") {
		return ConnectResult{Success: true}
	}

	return ConnectResult{Success: false, Error: output}
}

type InstallResult struct {
	Success bool
	Error   string
}

type ADBConfig struct {
	Serial  string
	APKPath string
}

func (c ADBConfig) Validate() error {
	if c.Serial == "" {
		return errors.New("device serial required")
	}
	if c.APKPath == "" {
		return errors.New("APK path required")
	}
	if _, err := os.Stat(c.APKPath); os.IsNotExist(err) {
		return errors.New("APK file not found: " + c.APKPath)
	}
	return nil
}

func (c ADBConfig) Install() InstallResult {
	if err := c.Validate(); err != nil {
		return InstallResult{Success: false, Error: err.Error()}
	}

	output, err := runAdb("-s", c.Serial, "install", "-r", c.APKPath)
	if err != nil {
		return InstallResult{Success: false, Error: err.Error()}
	}

	if strings.Contains(output, "Success") {
		return InstallResult{Success: true}
	}

	return InstallResult{Success: false, Error: output}
}

func (c ADBConfig) InstallMultiple(apkPaths []string) InstallResult {
	if len(apkPaths) == 0 {
		return InstallResult{Success: false, Error: "no APK paths provided"}
	}

	args := []string{"-s", c.Serial, "install-multiple", "-r"}
	args = append(args, apkPaths...)

	output, err := runAdb(args...)
	if err != nil {
		return InstallResult{Success: false, Error: err.Error()}
	}

	if strings.Contains(output, "Success") {
		return InstallResult{Success: true}
	}

	return InstallResult{Success: false, Error: output}
}

var adbServerStarted bool = false

func StartADBServer() error {
	output, err := runAdb("start-server")
	if err != nil {
		return err
	}
	_ = output
	adbServerStarted = true
	return nil
}

func StopADBServer() {
	if adbServerStarted {
		runAdb("kill-server")
		adbServerStarted = false
	}
}
