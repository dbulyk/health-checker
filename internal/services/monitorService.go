package services

import (
	"context"
	"errors"
	"github.com/yusufpapurcu/wmi"
	"health-checker/internal/configs"
	"health-checker/internal/models"
	"log/slog"
	"sync/atomic"
	"time"
)

const (
	NormalZone  = "normal"
	WarningZone = "warning"
	DangerZone  = "danger"
)

type proc struct {
	PercentProcessorTime uint64
	TimeStamp_Sys100NS   uint64
}

type mem struct {
	AvailableMBytes uint64
}

type memInfo struct {
	Capacity uint64
}

type Monitor struct {
	cpuUtilisation models.Utilisation
	ramUtilisation models.Utilisation
	PollCount      atomic.Int64
}

func NewMonitor() *Monitor {
	return &Monitor{}
}

func (m *Monitor) Start(ctx context.Context, cfg configs.Checker) {
	go func() {
		slog.Debug("monitoring of the processor load is started")

		err := m.GetCPUUtilisation(ctx, cfg.Interval)
		if err != nil {
			slog.Error("processor data retrieval error", "error", err)
		}
	}()

	go func() {
		slog.Debug("RAM load monitoring started")

		err := m.GetRAMUtilisation(ctx, cfg.Interval)
		if err != nil {
			slog.Error("RAM data retrieval error", "error", err)
		}
	}()
}

func (m *Monitor) GetCPUUtilisation(ctx context.Context, interval time.Duration) error {
	var (
		startPoint         []proc
		endPoint           []proc
		startPointProcTime uint64
		startPointTS       uint64
		endPointProcTime   uint64
		endPointTS         uint64
		highLoadCounter    int
	)

	const query = "SELECT * FROM Win32_PerfRawData_PerfOS_Processor WHERE Name = '_Total'"

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			err := wmi.Query(query, &startPoint)
			if err != nil {
				return err
			}

			if len(startPoint) == 0 {
				return errors.New("no processor data")
			}

			startPointProcTime = startPoint[0].PercentProcessorTime
			startPointTS = startPoint[0].TimeStamp_Sys100NS

			time.Sleep(interval)

			err = wmi.Query(query, &endPoint)
			if err != nil {
				return err
			}

			if len(endPoint) == 0 {
				return errors.New("no processor data")
			}

			endPointProcTime = endPoint[0].PercentProcessorTime
			endPointTS = endPoint[0].TimeStamp_Sys100NS

			/*
				CPU utilisation calculation mechanism
				is based on https://learn.microsoft.com/en-us/windows/win32/wmisdk/monitoring-performance-data#using-raw-performance-data-classes
			*/
			cpuUtil := (1.0 - float64(endPointProcTime-startPointProcTime)/float64(endPointTS-startPointTS)) * 100
			if cpuUtil > 75 {
				highLoadCounter++
			} else if highLoadCounter > 0 {
				highLoadCounter--
			}

			m.cpuUtilisation.Lock()

			if highLoadCounter > 10 || cpuUtil >= 90 {
				m.cpuUtilisation.LoadZone = DangerZone
			} else if highLoadCounter > 0 {
				m.cpuUtilisation.LoadZone = WarningZone
			} else {
				m.cpuUtilisation.LoadZone = NormalZone
			}

			m.cpuUtilisation.Value = cpuUtil

			slog.Debug("", "CPU load", cpuUtil)
			m.cpuUtilisation.Unlock()
		case <-ctx.Done():
			slog.Debug("CPU load monitoring is stopped")
			return nil
		}
	}
}

func (m *Monitor) GetRAMUtilisation(ctx context.Context, interval time.Duration) error {
	var (
		memI            []memInfo
		capacity        uint64
		availableMemory float64
		highLoadCounter int
	)

	query := "SELECT * FROM Win32_PhysicalMemory"
	err := wmi.Query(query, &memI)
	if err != nil {
		return err
	}
	if len(memI) == 0 {
		return errors.New("no memory data")
	}

	for _, v := range memI {
		capacity += v.Capacity
	}
	capacity = capacity / 1024 / 1024
	slog.Debug("", "memory capacity", capacity)

	query = "SELECT * FROM Win32_PerfFormattedData_PerfOS_Memory"
	var memoryPoint []mem

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			err = wmi.Query(query, &memoryPoint)
			if err != nil {
				return err
			}
			if len(memoryPoint) == 0 {
				return errors.New("no memory data")
			}

			availableMemory = float64(memoryPoint[0].AvailableMBytes) / float64(capacity) * 100
			slog.Debug("", "available memory in percent", availableMemory)

			if availableMemory < 25 {
				highLoadCounter++
			} else if highLoadCounter > 0 {
				highLoadCounter--
			}

			m.ramUtilisation.Lock()
			if highLoadCounter > 10 || availableMemory <= 10 {
				m.ramUtilisation.LoadZone = DangerZone
			} else if highLoadCounter > 0 {
				m.ramUtilisation.LoadZone = WarningZone
			} else {
				m.ramUtilisation.LoadZone = NormalZone
			}
			m.ramUtilisation.Value = availableMemory
			m.ramUtilisation.Unlock()
		case <-ctx.Done():
			slog.Debug("RAM load monitoring is stopped")
			return nil
		}
	}
}

func (m *Monitor) GetCPUUtilisationValue() *models.Utilisation {
	m.cpuUtilisation.Lock()
	defer m.cpuUtilisation.Unlock()

	return &m.cpuUtilisation
}

func (m *Monitor) GetRAMUtilisationValue() *models.Utilisation {
	m.ramUtilisation.Lock()
	defer m.ramUtilisation.Unlock()

	return &m.ramUtilisation
}
