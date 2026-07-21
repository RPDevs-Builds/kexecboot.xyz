package payload

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParseEndpointsYaml(t *testing.T) {
	mockYaml := `
endpoints:
  ubuntu-2404:
    path: /ubuntu/releases/24.04/
    os: ubuntu
    version: "24.04"
    files:
      - vmlinuz
      - initrd
  alpine-latest:
    path: /alpine/latest-stable/
    os: alpine
    version: "latest"
    files:
      - vmlinuz-virt
      - initramfs-virt
`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockYaml))
	}))
	defer server.Close()

	endpoints, err := FetchEndpoints(server.URL)
	if err != nil {
		t.Fatalf("Expected no error fetching endpoints, got %v", err)
	}

	if len(endpoints) != 2 {
		t.Fatalf("Expected 2 endpoints, got %d", len(endpoints))
	}

	foundUbuntu := false
	for _, ep := range endpoints {
		if ep.Name == "ubuntu-2404" {
			foundUbuntu = true
			if ep.Os != "ubuntu" {
				t.Errorf("Expected OS 'ubuntu', got '%s'", ep.Os)
			}
			if ep.Kernel == "" || ep.Initrd == "" {
				t.Errorf("Expected valid Kernel and Initrd URLs, got Kernel='%s', Initrd='%s'", ep.Kernel, ep.Initrd)
			}
		}
	}

	if !foundUbuntu {
		t.Errorf("Expected ubuntu-2404 endpoint in parsed list")
	}
}
