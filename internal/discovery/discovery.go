// Package discovery locates Google Cast receivers advertised over mDNS.
package discovery

import (
	"fmt"
	"net"
	"sort"
	"strings"
	"time"

	"github.com/hashicorp/mdns"
)

const service = "_googlecast._tcp"

// Device is the subset of a Google Cast mDNS advertisement needed by clients.
type Device struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Model           string `json:"model"`
	Host            string `json:"host"`
	Port            int    `json:"port"`
	ConnectionState string `json:"connectionState"`
}

// Discover searches for Cast receivers until timeout expires. Duplicate mDNS
// responses are collapsed and the result is sorted by friendly name.
func Discover(timeout time.Duration) ([]Device, error) {
	if timeout <= 0 {
		return nil, fmt.Errorf("discovery timeout must be positive")
	}

	entries := make(chan *mdns.ServiceEntry, 16)
	results := make(chan []Device, 1)

	go func() {
		seen := make(map[string]Device)
		for entry := range entries {
			device := deviceFromEntry(entry)
			key := device.ID
			if key == "" {
				key = net.JoinHostPort(device.Host, fmt.Sprint(device.Port))
			}
			seen[key] = device
		}

		result := make([]Device, 0, len(seen))
		for _, device := range seen {
			result = append(result, device)
		}
		sort.Slice(result, func(i, j int) bool {
			return strings.ToLower(result[i].Name) < strings.ToLower(result[j].Name)
		})
		results <- result
	}()

	params := mdns.DefaultParams(service)
	params.Entries = entries
	params.DisableIPv6 = true
	params.Timeout = timeout

	err := mdns.Query(params)
	close(entries)
	result := <-results
	if err != nil {
		return result, fmt.Errorf("discover cast devices: %w", err)
	}
	return result, nil
}

func deviceFromEntry(entry *mdns.ServiceEntry) Device {
	txt := make(map[string]string, len(entry.InfoFields))
	for _, field := range entry.InfoFields {
		key, value, ok := strings.Cut(field, "=")
		if ok {
			txt[key] = value
		}
	}

	host := strings.TrimSuffix(entry.Host, ".")
	if entry.AddrV4 != nil {
		host = entry.AddrV4.String()
	} else if entry.AddrV6 != nil {
		host = entry.AddrV6.String()
	}

	name := txt["fn"]
	if name == "" {
		name = strings.TrimSuffix(entry.Name, ".")
		if serviceIndex := strings.Index(name, "."+service); serviceIndex >= 0 {
			name = name[:serviceIndex]
		}
	}

	return Device{
		ID:              txt["id"],
		Name:            name,
		Model:           txt["md"],
		Host:            host,
		Port:            entry.Port,
		ConnectionState: "discovered",
	}
}
