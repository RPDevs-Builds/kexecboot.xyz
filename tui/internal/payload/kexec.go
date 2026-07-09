package payload

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"gopkg.in/yaml.v3"
)

// Endpoint represents a parsed payload endpoint ready for execution
type Endpoint struct {
	Name   string
	Os     string
	Kernel string // Full URL
	Initrd string // Full URL
	Append string // Kernel cmdline
}

// NetbootEndpoint represents the raw structure in netboot.xyz endpoints.yml
type NetbootEndpoint struct {
	Path    string   `yaml:"path"`
	Files   []string `yaml:"files"`
	Os      string   `yaml:"os"`
	Version string   `yaml:"version"`
	Flavor  string   `yaml:"flavor"`
}

// EndpointsResponse maps the top level endpoints map
type EndpointsResponse struct {
	Endpoints map[string]NetbootEndpoint `yaml:"endpoints"`
}

// FetchEndpoints retrieves and parses netboot.xyz endpoints.yml
func FetchEndpoints(url string) ([]Endpoint, error) {
	if url == "" {
		// Use jsdelivr to avoid GitHub raw 429 rate limits
		url = "https://cdn.jsdelivr.net/gh/netbootxyz/netboot.xyz@master/endpoints.yml"
	}
	
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch endpoints: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var data EndpointsResponse
	if err := yaml.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("failed to parse endpoints map: %w", err)
	}

	var endpoints []Endpoint
	baseURL := "https://github.com/netbootxyz"

	for key, raw := range data.Endpoints {
		ep := Endpoint{
			Name: key,
			Os:   raw.Os,
		}

		// Find kernel and initrd in the files list
		var vmlinuzName, initrdName string
		for _, f := range raw.Files {
			if strings.Contains(f, "vmlinuz") || strings.Contains(f, "bzImage") {
				vmlinuzName = f
			} else if strings.Contains(f, "initrd") {
				initrdName = f
			}
		}

		if vmlinuzName != "" && initrdName != "" {
			// Construct full download URLs
			ep.Kernel = baseURL + raw.Path + vmlinuzName
			ep.Initrd = baseURL + raw.Path + initrdName
			// Setup a basic cmdline (this is highly OS specific, but we'll add a generic one)
			ep.Append = "console=ttyS0,115200 console=tty0"
			endpoints = append(endpoints, ep)
		}
	}

	return endpoints, nil
}

// Execute downloads the payload and pivots the kernel using kexec
func Execute(e Endpoint) error {
	// 1. Download kernel and initrd to tmpfs
	fmt.Printf("\r\nDownloading Kernel from: %s\r\n", e.Kernel)
	if err := download(e.Kernel, "/tmp/vmlinuz"); err != nil {
		return fmt.Errorf("failed to download kernel: %w", err)
	}

	fmt.Printf("Downloading Initrd from: %s\r\n", e.Initrd)
	if err := download(e.Initrd, "/tmp/initrd"); err != nil {
		return fmt.Errorf("failed to download initrd: %w", err)
	}

	fmt.Printf("Loading kernel via kexec...\r\n")
	// 2. Load kernel via kexec
	cmd := exec.Command("kexec", "-l", "/tmp/vmlinuz", "--initrd=/tmp/initrd", "--append="+e.Append)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("kexec load failed: %w", err)
	}

	fmt.Printf("Pivoting kernel...\r\n")
	// 3. Execute kexec
	execCmd := exec.Command("kexec", "-e")
	return execCmd.Run()
}

func download(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
