package scanner

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/monify-labs/agent/pkg/models"
)

// PortScanner handles port scanning operations
type PortScanner struct {
	timeout    time.Duration
	maxWorkers int
}

// NewPortScanner creates a new port scanner
func NewPortScanner(timeout time.Duration, maxWorkers int) *PortScanner {
	return &PortScanner{
		timeout:    timeout,
		maxWorkers: maxWorkers,
	}
}

// Scan scans the specified ports on the target host
func (ps *PortScanner) Scan(ctx context.Context, target string, ports []int) (*models.PortScanResult, error) {
	startTime := time.Now()

	result := &models.PortScanResult{
		Target:    target,
		Timestamp: startTime,
		OpenPorts: make([]models.OpenPort, 0),
	}

	// Create a channel for ports to scan
	portChan := make(chan int, len(ports))
	for _, port := range ports {
		portChan <- port
	}
	close(portChan)

	// Create a channel for results
	resultChan := make(chan models.OpenPort, len(ports))

	// Start worker pool
	var wg sync.WaitGroup
	numWorkers := ps.maxWorkers
	if numWorkers > len(ports) {
		numWorkers = len(ports)
	}

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ps.scanWorker(ctx, target, portChan, resultChan)
		}()
	}

	// Wait for all workers to finish
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	for openPort := range resultChan {
		result.OpenPorts = append(result.OpenPorts, openPort)
	}

	result.ScanTime = time.Since(startTime)

	return result, nil
}

// scanWorker is a worker that scans ports from the port channel
func (ps *PortScanner) scanWorker(ctx context.Context, target string, portChan <-chan int, resultChan chan<- models.OpenPort) {
	for port := range portChan {
		select {
		case <-ctx.Done():
			return
		default:
			if ps.isPortOpen(ctx, target, port) {
				resultChan <- models.OpenPort{
					Port:    port,
					Service: getServiceName(port),
				}
			}
		}
	}
}

// isPortOpen checks if a port is open on the target
func (ps *PortScanner) isPortOpen(ctx context.Context, target string, port int) bool {
	address := fmt.Sprintf("%s:%d", target, port)

	// Create dialer with timeout
	dialer := &net.Dialer{
		Timeout: ps.timeout,
	}

	// Attempt to connect
	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return false
	}

	conn.Close()
	return true
}

// getServiceName returns the common service name for a port
func getServiceName(port int) string {
	commonPorts := map[int]string{
		20:    "ftp-data",
		21:    "ftp",
		22:    "ssh",
		23:    "telnet",
		25:    "smtp",
		53:    "dns",
		80:    "http",
		110:   "pop3",
		143:   "imap",
		443:   "https",
		465:   "smtps",
		587:   "smtp-submission",
		993:   "imaps",
		995:   "pop3s",
		3306:  "mysql",
		3389:  "rdp",
		5432:  "postgresql",
		5900:  "vnc",
		6379:  "redis",
		8080:  "http-proxy",
		8443:  "https-alt",
		9200:  "elasticsearch",
		27017: "mongodb",
	}

	if service, ok := commonPorts[port]; ok {
		return service
	}
	return ""
}

// ScanPortRange scans a range of ports
func (ps *PortScanner) ScanPortRange(ctx context.Context, target string, startPort, endPort int) (*models.PortScanResult, error) {
	ports := make([]int, 0, endPort-startPort+1)
	for port := startPort; port <= endPort; port++ {
		ports = append(ports, port)
	}
	return ps.Scan(ctx, target, ports)
}

// ScanCommonPorts scans commonly used ports
func (ps *PortScanner) ScanCommonPorts(ctx context.Context, target string) (*models.PortScanResult, error) {
	commonPorts := []int{
		21, 22, 23, 25, 53, 80, 110, 143, 443, 465, 587, 993, 995,
		3306, 3389, 5432, 5900, 6379, 8080, 8443, 9200, 27017,
	}
	return ps.Scan(ctx, target, commonPorts)
}
