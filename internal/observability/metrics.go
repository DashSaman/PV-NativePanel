package observability

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/sys/unix"
)

type RawSample struct {
	Timestamp            time.Time
	CPUIdle              uint64
	CPUTotal             uint64
	MemoryTotalBytes     uint64
	MemoryAvailableBytes uint64
	DiskTotalBytes       uint64
	DiskAvailableBytes   uint64
	Load1                float64
	Load5                float64
	Load15               float64
	UptimeSeconds        float64
	RXBytes              uint64
	TXBytes              uint64
	NetworkInterface     string
}

type Source interface {
	Read(context.Context) (RawSample, error)
}

type Snapshot struct {
	SampledAt            time.Time `json:"sampled_at"`
	CPUPercent           float64   `json:"cpu_percent"`
	MemoryTotalBytes     uint64    `json:"memory_total_bytes"`
	MemoryAvailableBytes uint64    `json:"memory_available_bytes"`
	MemoryUsedPercent    float64   `json:"memory_used_percent"`
	DiskTotalBytes       uint64    `json:"disk_total_bytes"`
	DiskAvailableBytes   uint64    `json:"disk_available_bytes"`
	DiskUsedPercent      float64   `json:"disk_used_percent"`
	Load1                float64   `json:"load_1"`
	Load5                float64   `json:"load_5"`
	Load15               float64   `json:"load_15"`
	UptimeSeconds        float64   `json:"uptime_seconds"`
	RXBytes              uint64    `json:"rx_bytes"`
	TXBytes              uint64    `json:"tx_bytes"`
	NetworkInterface     string    `json:"network_interface"`
	RateAvailable        bool      `json:"rate_available"`
	SampleWindowSeconds  float64   `json:"sample_window_seconds"`
	RXBytesPerSecond     float64   `json:"rx_bytes_per_second"`
	TXBytesPerSecond     float64   `json:"tx_bytes_per_second"`
}

type Collector struct {
	source   Source
	mu       sync.Mutex
	previous *RawSample
}

func NewCollector(source Source) *Collector {
	return &Collector{source: source}
}

func (c *Collector) Collect(ctx context.Context) (Snapshot, error) {
	if c == nil || c.source == nil {
		return Snapshot{}, errors.New("metrics source is unavailable")
	}
	raw, err := c.source.Read(ctx)
	if err != nil {
		return Snapshot{}, err
	}
	if raw.Timestamp.IsZero() {
		return Snapshot{}, errors.New("metrics sample timestamp is required")
	}

	result := snapshotFromRaw(raw)
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.previous != nil {
		window := raw.Timestamp.Sub(c.previous.Timestamp).Seconds()
		if window > 0 && raw.RXBytes >= c.previous.RXBytes && raw.TXBytes >= c.previous.TXBytes {
			result.RateAvailable = true
			result.SampleWindowSeconds = window
			result.RXBytesPerSecond = float64(raw.RXBytes-c.previous.RXBytes) / window
			result.TXBytesPerSecond = float64(raw.TXBytes-c.previous.TXBytes) / window
		}
		if raw.CPUTotal > c.previous.CPUTotal && raw.CPUIdle >= c.previous.CPUIdle {
			totalDelta := raw.CPUTotal - c.previous.CPUTotal
			idleDelta := raw.CPUIdle - c.previous.CPUIdle
			if idleDelta <= totalDelta {
				result.CPUPercent = 100 * float64(totalDelta-idleDelta) / float64(totalDelta)
			}
		}
	}
	copy := raw
	c.previous = &copy
	return result, nil
}

func snapshotFromRaw(raw RawSample) Snapshot {
	result := Snapshot{
		SampledAt:            raw.Timestamp.UTC(),
		MemoryTotalBytes:     raw.MemoryTotalBytes,
		MemoryAvailableBytes: raw.MemoryAvailableBytes,
		DiskTotalBytes:       raw.DiskTotalBytes,
		DiskAvailableBytes:   raw.DiskAvailableBytes,
		Load1:                raw.Load1,
		Load5:                raw.Load5,
		Load15:               raw.Load15,
		UptimeSeconds:        raw.UptimeSeconds,
		RXBytes:              raw.RXBytes,
		TXBytes:              raw.TXBytes,
		NetworkInterface:     raw.NetworkInterface,
	}
	if raw.CPUTotal > 0 && raw.CPUIdle <= raw.CPUTotal {
		result.CPUPercent = 100 * float64(raw.CPUTotal-raw.CPUIdle) / float64(raw.CPUTotal)
	}
	if raw.MemoryTotalBytes > 0 && raw.MemoryAvailableBytes <= raw.MemoryTotalBytes {
		result.MemoryUsedPercent = 100 * float64(raw.MemoryTotalBytes-raw.MemoryAvailableBytes) / float64(raw.MemoryTotalBytes)
	}
	if raw.DiskTotalBytes > 0 && raw.DiskAvailableBytes <= raw.DiskTotalBytes {
		result.DiskUsedPercent = 100 * float64(raw.DiskTotalBytes-raw.DiskAvailableBytes) / float64(raw.DiskTotalBytes)
	}
	return result
}

type LinuxSource struct {
	Now func() time.Time
}

func NewLinuxSource() *LinuxSource {
	return &LinuxSource{Now: time.Now}
}

func (s *LinuxSource) Read(ctx context.Context) (RawSample, error) {
	select {
	case <-ctx.Done():
		return RawSample{}, ctx.Err()
	default:
	}
	cpuIdle, cpuTotal, err := readCPUStat("/proc/stat")
	if err != nil {
		return RawSample{}, err
	}
	memoryTotal, memoryAvailable, err := readMemory("/proc/meminfo")
	if err != nil {
		return RawSample{}, err
	}
	load1, load5, load15, err := readLoad("/proc/loadavg")
	if err != nil {
		return RawSample{}, err
	}
	uptime, err := readUptime("/proc/uptime")
	if err != nil {
		return RawSample{}, err
	}
	rxBytes, txBytes, interfaces, err := readNetwork("/proc/net/dev")
	if err != nil {
		return RawSample{}, err
	}
	var stat unix.Statfs_t
	if err := unix.Statfs("/", &stat); err != nil {
		return RawSample{}, fmt.Errorf("read root filesystem: %w", err)
	}
	blockSize := uint64(stat.Bsize)
	now := time.Now
	if s != nil && s.Now != nil {
		now = s.Now
	}
	return RawSample{
		Timestamp:            now().UTC(),
		CPUIdle:              cpuIdle,
		CPUTotal:             cpuTotal,
		MemoryTotalBytes:     memoryTotal,
		MemoryAvailableBytes: memoryAvailable,
		DiskTotalBytes:       stat.Blocks * blockSize,
		DiskAvailableBytes:   stat.Bavail * blockSize,
		Load1:                load1,
		Load5:                load5,
		Load15:               load15,
		UptimeSeconds:        uptime,
		RXBytes:              rxBytes,
		TXBytes:              txBytes,
		NetworkInterface:     interfaces,
	}, nil
}

func readCPUStat(path string) (uint64, uint64, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, 0, fmt.Errorf("read cpu metrics: %w", err)
	}
	defer file.Close()
	line, err := bufio.NewReader(file).ReadString('\n')
	if err != nil {
		return 0, 0, fmt.Errorf("read cpu metrics: %w", err)
	}
	fields := strings.Fields(line)
	if len(fields) < 5 || fields[0] != "cpu" {
		return 0, 0, errors.New("unexpected /proc/stat cpu format")
	}
	var values []uint64
	for _, field := range fields[1:] {
		value, parseErr := strconv.ParseUint(field, 10, 64)
		if parseErr != nil {
			return 0, 0, errors.New("invalid /proc/stat cpu value")
		}
		values = append(values, value)
	}
	var total uint64
	for _, value := range values {
		total += value
	}
	idle := values[3]
	if len(values) > 4 {
		idle += values[4]
	}
	return idle, total, nil
}

func readMemory(path string) (uint64, uint64, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, 0, fmt.Errorf("read memory metrics: %w", err)
	}
	defer file.Close()
	var total, available uint64
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 2 {
			continue
		}
		value, parseErr := strconv.ParseUint(fields[1], 10, 64)
		if parseErr != nil {
			continue
		}
		switch fields[0] {
		case "MemTotal:":
			total = value * 1024
		case "MemAvailable:":
			available = value * 1024
		}
	}
	if err := scanner.Err(); err != nil {
		return 0, 0, fmt.Errorf("read memory metrics: %w", err)
	}
	if total == 0 {
		return 0, 0, errors.New("MemTotal is unavailable")
	}
	return total, available, nil
}

func readLoad(path string) (float64, float64, float64, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("read load metrics: %w", err)
	}
	fields := strings.Fields(string(data))
	if len(fields) < 3 {
		return 0, 0, 0, errors.New("unexpected /proc/loadavg format")
	}
	values := make([]float64, 3)
	for i := range values {
		values[i], err = strconv.ParseFloat(fields[i], 64)
		if err != nil {
			return 0, 0, 0, errors.New("invalid /proc/loadavg value")
		}
	}
	return values[0], values[1], values[2], nil
}

func readUptime(path string) (float64, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, fmt.Errorf("read uptime: %w", err)
	}
	fields := strings.Fields(string(data))
	if len(fields) == 0 {
		return 0, errors.New("unexpected /proc/uptime format")
	}
	value, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return 0, errors.New("invalid /proc/uptime value")
	}
	return value, nil
}

func readNetwork(path string) (uint64, uint64, string, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, 0, "", fmt.Errorf("read network metrics: %w", err)
	}
	defer file.Close()
	var rx, tx uint64
	var names []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		colon := strings.IndexByte(line, ':')
		if colon < 1 {
			continue
		}
		name := strings.TrimSpace(line[:colon])
		if name == "lo" {
			continue
		}
		fields := strings.Fields(line[colon+1:])
		if len(fields) < 9 {
			continue
		}
		rxValue, rxErr := strconv.ParseUint(fields[0], 10, 64)
		txValue, txErr := strconv.ParseUint(fields[8], 10, 64)
		if rxErr != nil || txErr != nil {
			continue
		}
		rx += rxValue
		tx += txValue
		names = append(names, name)
	}
	if err := scanner.Err(); err != nil {
		return 0, 0, "", fmt.Errorf("read network metrics: %w", err)
	}
	return rx, tx, strings.Join(names, ","), nil
}
