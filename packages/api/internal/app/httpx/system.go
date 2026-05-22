package httpx

import (
	"math"
	"net/http"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type SystemStats struct {
	CPUPercent    float64 `json:"cpuPercent"`
	MemoryUsedGB  float64 `json:"memoryUsedGB"`
	MemoryTotalGB float64 `json:"memoryTotalGB"`
	LoadAvg1      float64 `json:"loadAvg1"`
}

func (s Server) getSystemStats(w http.ResponseWriter, _ *http.Request) {
	totalGB, usedBytes := readMemory()

	stats := SystemStats{
		CPUPercent:    cachedCPUPercent(),
		MemoryTotalGB: totalGB,
		MemoryUsedGB:  bytesToGB(usedBytes),
		LoadAvg1:      readLoadAvg1(),
	}

	writeJSON(w, http.StatusOK, stats)
}

// CPU sampling on macOS requires two `top` snapshots ~1s apart and
// blocks ~3.5s. Sample in the background; the handler reads the latest.
var (
	cpuSampleOnce sync.Once
	cpuPercent    atomic.Uint64 // float64 bits
)

func cachedCPUPercent() float64 {
	cpuSampleOnce.Do(startCPUSampler)

	return math.Float64frombits(cpuPercent.Load())
}

func startCPUSampler() {
	go func() {
		for {
			cpuPercent.Store(math.Float64bits(readCPUPercent()))

			time.Sleep(500 * time.Millisecond)
		}
	}()
}

var cpuUsageRegex = regexp.MustCompile(`CPU usage:\s*([\d.]+)%\s*user,\s*([\d.]+)%\s*sys`)

func readCPUPercent() float64 {
	out, err := exec.Command("top", "-l", "2", "-n", "0", "-s", "1").Output()

	if err != nil {
		return 0
	}

	matches := cpuUsageRegex.FindAllStringSubmatch(string(out), -1)

	if len(matches) == 0 {
		return 0
	}

	last := matches[len(matches)-1]

	if len(last) < 3 {
		return 0
	}

	user, _ := strconv.ParseFloat(last[1], 64)
	sys, _ := strconv.ParseFloat(last[2], 64)

	return user + sys
}

var (
	vmStatPageSizeRegex = regexp.MustCompile(`page size of (\d+) bytes`)
	vmStatLineRegex     = regexp.MustCompile(`^([^:]+):\s+(\d+)\.?\s*$`)
)

func readMemory() (totalGB float64, usedBytes uint64) {
	totalGB = readMemoryTotalGB()

	out, err := exec.Command("vm_stat").Output()

	if err != nil {
		return totalGB, 0
	}

	text := string(out)
	pageBytes := uint64(4096)

	if m := vmStatPageSizeRegex.FindStringSubmatch(text); len(m) == 2 {
		if v, err := strconv.ParseUint(m[1], 10, 64); err == nil && v > 0 {
			pageBytes = v
		}
	}

	var active, wired, compressed uint64

	for _, line := range strings.Split(text, "\n") {
		m := vmStatLineRegex.FindStringSubmatch(strings.TrimSpace(line))

		if len(m) != 3 {
			continue
		}

		label := strings.TrimSpace(m[1])
		value, err := strconv.ParseUint(m[2], 10, 64)

		if err != nil {
			continue
		}

		switch label {
		case "Pages active":
			active = value
		case "Pages wired down":
			wired = value
		case "Pages occupied by compressor":
			compressed = value
		}
	}

	usedBytes = (active + wired + compressed) * pageBytes

	return totalGB, usedBytes
}

func readMemoryTotalGB() float64 {
	out, err := exec.Command("sysctl", "-n", "hw.memsize").Output()

	if err != nil {
		return 0
	}

	bytes, err := strconv.ParseUint(strings.TrimSpace(string(out)), 10, 64)

	if err != nil {
		return 0
	}

	return bytesToGB(bytes)
}

func readLoadAvg1() float64 {
	out, err := exec.Command("sysctl", "-n", "vm.loadavg").Output()

	if err != nil {
		return 0
	}

	fields := strings.Fields(strings.Trim(strings.TrimSpace(string(out)), "{}"))

	if len(fields) == 0 {
		return 0
	}

	v, err := strconv.ParseFloat(fields[0], 64)

	if err != nil {
		return 0
	}

	return v
}

func bytesToGB(b uint64) float64 {
	return float64(b) / (1 << 30)
}
