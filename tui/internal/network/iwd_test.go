package network

import (
	"reflect"
	"testing"
)

func TestParseNetworksOutput(t *testing.T) {
	mockIwctlOutput := `
                               Wait for network scanning...                           
  Network name                         Security          Signal
  -------------------------------------------------------------------------------
  > RPDevs-5G                           psk               ****
    Home-WiFi                          psk               ***
    Guest-Access                       open              **
`

	expected := []string{"RPDevs-5G", "Home-WiFi", "Guest-Access"}
	result := ParseNetworksOutput(mockIwctlOutput)

	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("Expected %v, got %v", expected, result)
	}
}
