package network

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

func getDevice() (string, error) {
	cmd := exec.Command("iwctl", "device", "list")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to list devices: %w", err)
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if strings.Contains(line, "station") {
			fields := strings.Fields(line)
			if len(fields) > 1 {
				for _, word := range fields {
					if word == "station" {
						// Usually the device name is the first field
						return fields[0], nil
					}
				}
			}
		}
	}
	return "wlan0", nil // default fallback
}

// Scan returns a list of visible SSIDs
func Scan() ([]string, error) {
	dev, err := getDevice()
	if err != nil {
		dev = "wlan0"
	}

	// Trigger a scan
	exec.Command("iwctl", "station", dev, "scan").Run()

	cmd := exec.Command("iwctl", "station", dev, "get-networks")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get networks: %w", err)
	}

	ssids := ParseNetworksOutput(string(out))
	if len(ssids) == 0 {
		return nil, fmt.Errorf("no networks found")
	}

	return ssids, nil
}

// ParseNetworksOutput parses iwctl station get-networks text output
func ParseNetworksOutput(out string) []string {
	var ssids []string
	lines := strings.Split(out, "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		ansiRegex := regexp.MustCompile(`\x1b\[[0-9;]*[mK]`)
		line = ansiRegex.ReplaceAllString(line, "")
		
		if line == "" || strings.HasPrefix(line, "-") || strings.Contains(line, "Network name") || strings.Contains(line, "Available networks") || strings.Contains(line, "scanning") {
			continue
		}
		
		line = strings.TrimPrefix(line, ">")
		line = strings.TrimSpace(line)
		
		parts := regexp.MustCompile(`\s{2,}`).Split(line, -1)
		var ssid string
		if len(parts) > 0 {
			ssid = parts[0]
		}

		if ssid != "" {
			found := false
			for _, s := range ssids {
				if s == ssid {
					found = true
					break
				}
			}
			if !found {
				ssids = append(ssids, ssid)
			}
		}
	}
	return ssids
}

// Connect attempts to connect to a network with the given PSK
func Connect(ssid, psk string) error {
	dev, err := getDevice()
	if err != nil {
		dev = "wlan0"
	}

	var cmd *exec.Cmd
	if psk == "" {
		cmd = exec.Command("iwctl", "station", dev, "connect", ssid)
	} else {
		cmd = exec.Command("iwctl", "--passphrase", psk, "station", dev, "connect", ssid)
	}

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to connect: %w, stderr: %s", err, stderr.String())
	}
	return nil
}
