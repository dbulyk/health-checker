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

type net struct {
	CurrentBandwidth uint32
	BytesTotalPerSec uint64
}

type Monitor struct {
	cpuUtilization models.Utilization
	ramUtilization models.Utilization
	netUtilization models.Utilization
	PollCount      atomic.Int64
}

func NewMonitor() *Monitor {
	return &Monitor{}
}

func (m *Monitor) Start(ctx context.Context, cfg configs.Checker) {
	go func() {
		slog.Debug("monitoring of the processor load is started")

		err := m.GetCPUUtilization(ctx, cfg.Interval)
		if err != nil {
			slog.Error("processor data retrieval error", "error", err)
		}
	}()

	go func() {
		slog.Debug("RAM load monitoring started")

		err := m.GetRAMUtilization(ctx, cfg.Interval)
		if err != nil {
			slog.Error("RAM data retrieval error", "error", err)
		}
	}()

	go func() {
		slog.Debug("network load monitoring started")

		err := m.getNetUtilization(ctx, cfg.Interval)
		if err != nil {
			slog.Error("network data retrieval error", "error", err)
		}
	}()
}

func (m *Monitor) GetCPUUtilization(ctx context.Context, interval time.Duration) error {
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
				CPU utilization calculation mechanism
				is based on https://learn.microsoft.com/en-us/windows/win32/wmisdk/monitoring-performance-data#using-raw-performance-data-classes
			*/
			cpuUtil := (1.0 - float64(endPointProcTime-startPointProcTime)/float64(endPointTS-startPointTS)) * 100
			if cpuUtil > 75 {
				highLoadCounter++
			} else if highLoadCounter > 0 {
				highLoadCounter--
			}

			m.cpuUtilization.Lock()

			if highLoadCounter > 10 || cpuUtil >= 90 {
				m.cpuUtilization.LoadZone = DangerZone
			} else if highLoadCounter > 0 {
				m.cpuUtilization.LoadZone = WarningZone
			} else {
				m.cpuUtilization.LoadZone = NormalZone
			}

			m.cpuUtilization.Value = cpuUtil

			slog.Debug("", "CPU load", cpuUtil)
			m.cpuUtilization.Unlock()
		case <-ctx.Done():
			slog.Debug("CPU load monitoring is stopped")
			return nil
		}
	}
}

func (m *Monitor) GetRAMUtilization(ctx context.Context, interval time.Duration) error {
	type memInfo struct {
		Capacity uint64
	}

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

			m.ramUtilization.Lock()
			if highLoadCounter > 10 || availableMemory <= 10 {
				m.ramUtilization.LoadZone = DangerZone
			} else if highLoadCounter > 0 {
				m.ramUtilization.LoadZone = WarningZone
			} else {
				m.ramUtilization.LoadZone = NormalZone
			}
			m.ramUtilization.Value = availableMemory
			m.ramUtilization.Unlock()
		case <-ctx.Done():
			slog.Debug("RAM load monitoring is stopped")
			return nil
		}
	}
}

func (m *Monitor) getNetUtilization(ctx context.Context, interval time.Duration) error {
	var (
		netInfo         []net
		highLoadCounter int
		netUtil         float64
	)
	query := "SELECT * FROM Win32_PerfFormattedData_Tcpip_NetworkInterface"
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			err := wmi.Query(query, &netInfo)
			if err != nil {
				return err
			}
			if len(netInfo) == 0 {
				return errors.New("no network data")
			}

			netUtil = float64(netInfo[0].BytesTotalPerSec) / float64(netInfo[0].CurrentBandwidth) * 1000
			slog.Debug("", "network utilization", netUtil)

			if netUtil > 85 {
				highLoadCounter++
			} else if highLoadCounter > 0 {
				highLoadCounter--
			}

			m.netUtilization.Lock()
			if highLoadCounter > 10 || netUtil >= 95 {
				m.netUtilization.LoadZone = DangerZone
			} else if highLoadCounter > 0 {
				m.netUtilization.LoadZone = WarningZone
			} else {
				m.netUtilization.LoadZone = NormalZone
			}
			m.netUtilization.Value = netUtil
			m.netUtilization.Unlock()
		case <-ctx.Done():
			slog.Debug("network load monitoring is stopped")
			return nil
		}
	}

}

func (m *Monitor) GetCPUUtilizationValue() *models.Utilization {
	return &m.cpuUtilization
}

func (m *Monitor) GetRAMUtilizationValue() *models.Utilization {
	return &m.ramUtilization
}

func (m *Monitor) GetNetUtilizationValue() *models.Utilization {
	return &m.netUtilization
}
