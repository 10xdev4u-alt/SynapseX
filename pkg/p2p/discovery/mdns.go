package discovery

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/grandcat/zeroconf"
)

// ServiceName is the mDNS service name for Synapse nodes
const ServiceName = "_synapse._tcp"

// Peer represents a discovered peer
type Peer struct {
	ID       string
	Address  string
	Port     int
	Hostname string
	TTL      time.Duration
}

// MDNSDiscoverer handles mDNS-based peer discovery
type MDNSDiscoverer struct {
	serviceName string
	domain      string
	instance    string
	port        int
	txtRecords  []string
	server      *zeroconf.Server
	stopCh      chan struct{}
}

// NewMDNSDiscoverer creates a new mDNS discoverer
func NewMDNSDiscoverer(instance string, port int, txtRecords []string) *MDNSDiscoverer {
	return &MDNSDiscoverer{
		serviceName: ServiceName,
		domain:      "local.",
		instance:    instance,
		port:        port,
		txtRecords:  txtRecords,
		stopCh:      make(chan struct{}),
	}
}

// Start begins advertising the service and discovering peers
func (m *MDNSDiscoverer) Start(ctx context.Context) error {
	// Start the mDNS server to advertise our service
	server, err := zeroconf.Register(m.instance, m.serviceName, m.domain, m.port, m.txtRecords, nil)
	if err != nil {
		return fmt.Errorf("failed to register mDNS service: %w", err)
	}
	m.server = server

	// Start discovery in a separate goroutine
	go m.discover(ctx)

	return nil
}

// Stop stops the mDNS discovery and advertising
func (m *MDNSDiscoverer) Stop() {
	if m.server != nil {
		m.server.Shutdown()
	}
	close(m.stopCh)
}

// discover continuously looks for other Synapse nodes on the network
func (m *MDNSDiscoverer) discover(ctx context.Context) {
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		log.Printf("Failed to create mDNS resolver: %v", err)
		return
	}

	entries := make(chan *zeroconf.ServiceEntry)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case entry := <-entries:
				// Process discovered peer
				peer := m.processEntry(entry)
				if peer != nil {
					// TODO: Handle discovered peer (send to main network)
					log.Printf("Discovered peer: %+v", peer)
				}
			}
		}
	}()

	ctx2, cancel := context.WithCancel(ctx)
	go func() {
		<-m.stopCh
		cancel()
	}()

	err = resolver.Browse(ctx2, m.serviceName, m.domain, entries)
	if err != nil {
		log.Printf("Failed to browse for mDNS services: %v", err)
	}
}

// processEntry converts a service entry to a Peer
func (m *MDNSDiscoverer) processEntry(entry *zeroconf.ServiceEntry) *Peer {
	if len(entry.AddrIPv4) == 0 && len(entry.AddrIPv6) == 0 {
		return nil
	}

	var address string
	if len(entry.AddrIPv4) > 0 {
		address = entry.AddrIPv4[0].String()
	} else {
		address = entry.AddrIPv6[0].String()
	}

	// Extract node ID from TXT records if available
	var nodeID string
	for _, txt := range entry.Text {
		if strings.HasPrefix(txt, "node_id=") {
			nodeID = strings.TrimPrefix(txt, "node_id=")
			break
		}
	}

	return &Peer{
		ID:       nodeID,
		Address:  address,
		Port:     entry.Port,
		Hostname: entry.HostName,
		TTL:      time.Duration(entry.TTL) * time.Second,
	}
}

// GetLocalIPs returns all local IP addresses
func GetLocalIPs() ([]string, error) {
	var ips []string
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, i := range interfaces {
		byt, err := i.Addrs()
		if err != nil {
			continue
		}
		for _, b := range byt {
			if ipnet, ok := b.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					ips = append(ips, ipnet.IP.String())
				}
			}
		}
	}
	return ips, nil
}
