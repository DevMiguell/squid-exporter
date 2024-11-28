package collector

import (
	"bufio"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

// Define the metrics for memory pools
var (
	memPoolObjSize = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "squid_mempool_obj_size_bytes",
			Help: "Size of each object in the pool in bytes",
		},
		[]string{"k_id", "pool"},
	)
	memPoolChunks = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "squid_mempool_chunks_kb_per_chunk",
			Help: "Chunk size in kilobytes",
		},
		[]string{"k_id", "pool"},
	)
)

// MemPoolCollector collects memory pool metrics from Squid
type MemPoolCollector struct {
	hostname string
	port     int
	client   *http.Client
}

// NewMemPoolCollector initializes the collector
func NewMemPoolCollector(hostname string, port int) *MemPoolCollector {
	return &MemPoolCollector{
		hostname: hostname,
		port:     port,
		client:   &http.Client{},
	}
}

// Describe registers the metrics descriptions
func (c *MemPoolCollector) Describe(ch chan<- *prometheus.Desc) {
	memPoolObjSize.Describe(ch)
	memPoolChunks.Describe(ch)
}

// Collect fetches the data from Squid and sets the metrics
func (c *MemPoolCollector) Collect(ch chan<- prometheus.Metric) {
	url := fmt.Sprintf("http://%s:%d/squid-internal-mgr/mem", c.hostname, c.port)

	resp, err := c.client.Get(url)
	if err != nil {
		fmt.Printf("Error fetching data from Squid: %v\n", err)
		return
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()

		// Example parsing logic for lines in mem.txt
		// Assume line format: "kid1 pool_name value"
		if strings.HasPrefix(line, "kid") {
			parts := strings.Fields(line)
			if len(parts) < 3 {
				continue
			}

			kid := parts[0]
			pool := parts[1]
			value, err := strconv.ParseFloat(parts[2], 64)
			if err != nil {
				fmt.Printf("Error parsing value: %v\n", err)
				continue
			}

			// Set the metrics
			memPoolObjSize.WithLabelValues(kid, pool).Set(value)
			// For demonstration, we set chunks as a derived value
			memPoolChunks.WithLabelValues(kid, pool).Set(value / 1024)
		}
	}

	// Push metrics to Prometheus
	memPoolObjSize.Collect(ch)
	memPoolChunks.Collect(ch)
}
