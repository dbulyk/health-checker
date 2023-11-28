package services

import (
	"context"
	"errors"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"health-checker/internal/configs"
	"health-checker/internal/models"
	"log/slog"
	"strconv"
	"sync/atomic"
	"time"
)

type Monitor struct {
	cpuUtilization models.Utilization
	ramUtilization models.Utilization
	PollCount      atomic.Int64
}

func NewMonitor() *Monitor {
	return &Monitor{}
}

func (m *Monitor) Start(ctx context.Context, cfg configs.Checker) {
	go func() {
		slog.Debug("cpu load started")

		err := m.GetCPUUtilization(ctx, cfg.Interval)
		if err != nil {
			slog.Error("cpu load error", "error", err)
		}
	}()

	go func() {
		slog.Debug("ram load started")

		err := m.GetRAMUtilization(ctx, cfg.Interval)
		if err != nil {
			slog.Error("ram load error", "error", err)
		}
	}()
}

func (m *Monitor) GetCPUUtilization(ctx context.Context, interval time.Duration) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			percentage, err := cpu.PercentWithContext(ctx, interval, false)
			if err != nil && errors.Is(err, context.Canceled) {
				return nil
			} else if err != nil {
				return err
			}

			m.cpuUtilization.Lock()
			m.cpuUtilization.Percentages = percentage[0]
			m.cpuUtilization.Unlock()

			slog.Debug("cpu load", "load", strconv.FormatFloat(m.cpuUtilization.Percentages, 'f', 2, 64))
		case <-ctx.Done():
			slog.Debug("cpu load update stopped")
			return nil
		}
	}
}

func (m *Monitor) GetRAMUtilization(ctx context.Context, interval time.Duration) error {
	memoryInfo, err := mem.VirtualMemory()
	if err != nil {
		return err
	}
	memoryUsage := memoryInfo.UsedPercent

	m.PollCount.Store(1)

	m.ramUtilization.Lock()
	m.ramUtilization.Percentages = memoryUsage
	m.ramUtilization.Unlock()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			memoryInfo, err = mem.VirtualMemoryWithContext(ctx)
			if err != nil {
				return err
			}
			memoryUsage = memoryInfo.UsedPercent

			m.ramUtilization.Lock()
			if m.PollCount.Load() > 5 {
				m.ramUtilization.Percentages = memoryUsage
				m.PollCount.Swap(1)
			} else {
				m.ramUtilization.Percentages += memoryUsage
				m.PollCount.Add(1)
			}
			m.ramUtilization.Unlock()
			slog.Debug("memory utilization", "utilization", strconv.FormatFloat(memoryUsage, 'f', 2, 64))
		case <-ctx.Done():
			slog.Debug("memory utilization update stopped")
			return nil
		}
	}
}

func (m *Monitor) GetCPUUtilizationValue() float64 {
	m.cpuUtilization.Lock()
	defer m.cpuUtilization.Unlock()

	return m.cpuUtilization.Percentages
}

func (m *Monitor) GetRAMUtilizationValue() float64 {
	m.ramUtilization.Lock()
	defer m.ramUtilization.Unlock()

	return m.ramUtilization.Percentages / float64(m.PollCount.Load())
}
