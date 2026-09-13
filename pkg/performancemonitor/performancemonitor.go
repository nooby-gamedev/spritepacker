package performancemonitor

import (
	"encoding/json"
	"os"
	"time"

	"github.com/rs/zerolog/log"
)

type AverageDeltaTime struct {
	Total   float64 `json:"total_delta_time"`
	Counter int     `json:"counter"`
	Avg     float64 `json:"average_delta_time"`
}
type PerformanceMonitor struct {
	AvgDelta map[string]AverageDeltaTime `json:"avg_delta"`

	monitorAvgDelta map[string]time.Time
	enabled         bool
	counterLimit    int
}

var session *PerformanceMonitor

// Returns a PerformanceMonitor using Singleton pattern.
// If already initialized, returns the existing instance.
// Otherwise, returns a new DISABLED Monitor.
//
// Use Enable() to enable the monitor, or use NewEnabled().
func Monitor() *PerformanceMonitor {
	if session == nil {
		session = &PerformanceMonitor{
			AvgDelta:        make(map[string]AverageDeltaTime),
			monitorAvgDelta: make(map[string]time.Time),
		}
	}
	return session
}

// Set counter limit (0 = unlimited).
func (p *PerformanceMonitor) SetCounterLimit(value int) *PerformanceMonitor {
	p.counterLimit = value
	return p
}
func (p *PerformanceMonitor) Enable() *PerformanceMonitor {
	p.enabled = true
	return p
}
func (p *PerformanceMonitor) Disable() *PerformanceMonitor {
	p.enabled = false
	return p
}
func (p *PerformanceMonitor) StartMeasureAverageDeltaTime(key string) {
	if !p.enabled {
		return
	}
	_, ok := p.monitorAvgDelta[key]
	if ok {
		log.Fatal().Str("key", key).Msg("fatal error: cannot start measuring the same key simultaneously")
	}

	p.monitorAvgDelta[key] = time.Now()
}
func (p *PerformanceMonitor) StopMeasureAverageDeltaTime(key string) {
	if !p.enabled {
		return
	}
	start, ok := p.monitorAvgDelta[key]
	if !ok {
		log.Fatal().Str("key", key).Msg("fatal error: cannot stop measuring the key because it wasn't being monitored")
	}
	now := time.Now()
	diff := now.Sub(start).Seconds()

	avg, ok := p.AvgDelta[key]
	if !ok {
		avg = AverageDeltaTime{}
	}

	if p.counterLimit > 0 {
		if avg.Counter >= p.counterLimit {
			delete(p.monitorAvgDelta, key)
			return
		}
	}

	avg.Counter++
	avg.Total += diff
	avg.Avg = avg.Total / float64(avg.Counter)
	p.AvgDelta[key] = avg
	delete(p.monitorAvgDelta, key)
}

// Save as JSON
func (p *PerformanceMonitor) SaveJson(savePath string) error {
	handle, err := os.Create(savePath)
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(handle)
	encoder.SetIndent("", "	")
	if err := encoder.Encode(p); err != nil {
		return err
	}
	return nil
}

// cache: 0000027963092590861513
// nocac: 00003676508300395364
