package client

import (
	"io"
	"log"
	"strings"

	"github.com/hashicorp/mdns"
)

type DirigeraHub struct {
	HostName        string
	Address         string
	Port            int
	SerialNumber    string
	FirmwareVersion string
}

// Scan searches for IKEA Smart-Home hubs in the network using mDNS.
func Scan() ([]DirigeraHub, error) {
	var hubs []DirigeraHub

	// Disable log output temporary
	oldLog := log.Default()
	log.SetOutput(io.Discard)
	defer log.SetOutput(oldLog.Writer())

	// Scan for hubs
	entriesChannel := make(chan *mdns.ServiceEntry, 4)
	go func() {
		for entry := range entriesChannel {
			info := convertToMap(entry.InfoFields)
			if info["type"] == "DIRIGERA" {
				hubs = append(hubs, DirigeraHub{
					HostName:        info["hostname"],
					Address:         entry.AddrV4.String(),
					Port:            entry.Port,
					FirmwareVersion: info["sv"],
					SerialNumber:    info["uuid"],
				})

			}
		}
	}()
	params := mdns.DefaultParams("_ihsp._tcp")
	params.Entries = entriesChannel
	params.DisableIPv6 = true
	err := mdns.Query(params)
	close(entriesChannel)

	return hubs, err
}

func convertToMap(array []string) map[string]string {
	result := make(map[string]string)
	for _, entry := range array {
		parts := strings.SplitN(entry, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := parts[0]
		value := parts[1]
		result[key] = value
	}
	return result
}
