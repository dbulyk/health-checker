package services

import (
	"context"
	"errors"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/yusufpapurcu/wmi"
	"health-checker/internal/configs"
	"health-checker/internal/models"
	"log/slog"
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

type disk struct {
	PercentDiskTime uint64
}

type networkName struct {
	InterfaceDescription string
}

type Monitor struct {
	cpuUtilization  models.Utilization
	ramUtilization  models.Utilization
	netUtilization  models.Utilization
	diskUtilization models.Utilization
}

var (
	cpu = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "cpu_utilization",
			Help: "Утилизация процессора",
		})

	memory = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "memory_utilization",
			Help: "Утилизация оперативной памяти",
		})

	diskIO = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "disk_utilization",
			Help: "Утилизация I/O диска",
		})

	network = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "network_utilization",
			Help: "Утилизация сети",
		})
)

func NewMonitor() *Monitor {
	return &Monitor{}
}

func (m *Monitor) Start(ctx context.Context, cfg configs.Checker) {
	prometheus.MustRegister(cpu, memory, diskIO, network)
	go func() {
		slog.Debug("monitoring of the processor load is started")

		err := m.getCPUUtilization(ctx, cfg.Interval)
		if err != nil {
			slog.Error("processor data retrieval error", "error", err)
		}
	}()

	go func() {
		slog.Debug("RAM load monitoring started")

		err := m.getRAMUtilization(ctx, cfg.Interval)
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

	go func() {
		slog.Debug("disk load monitoring started")

		err := m.getDiskUtilization(ctx, cfg.Interval)
		if err != nil {
			slog.Error("disk data retrieval error", "error", err)
		}
	}()
}

func (m *Monitor) getCPUUtilization(ctx context.Context, interval time.Duration) error {
	var (
		startPoint         []proc
		endPoint           []proc
		startPointProcTime uint64
		startPointTS       uint64
		endPointProcTime   uint64
		endPointTS         uint64
		cpuUtil            float64
		highLoadCounter    int
		cpuUtilFormatted   string
		err                error
	)

	const query = "SELECT PercentProcessorTime, TimeStamp_Sys100NS FROM Win32_PerfRawData_PerfOS_Processor WHERE Name = '_Total'"

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			err = wmi.Query(query, &startPoint)
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
			cpuUtil = (1.0 - float64(endPointProcTime-startPointProcTime)/float64(endPointTS-startPointTS)) * 100
			if cpuUtil >= 75 {
				highLoadCounter++
			} else if highLoadCounter > 0 {
				highLoadCounter--
			}
			cpuUtilFormatted = fmt.Sprintf("%.*f", 2, cpuUtil)

			m.cpuUtilization.Lock()

			if cpuUtil >= 90 {
				m.cpuUtilization.LoadZone = DangerZone
			} else if highLoadCounter >= 10 {
				m.cpuUtilization.LoadZone = WarningZone
			} else {
				m.cpuUtilization.LoadZone = NormalZone
			}

			m.cpuUtilization.Value = cpuUtilFormatted
			m.cpuUtilization.Unlock()
			cpu.Set(cpuUtil)

			slog.Debug("", "CPU load", cpuUtilFormatted)
		case <-ctx.Done():
			slog.Debug("CPU load monitoring is stopped")
			return nil
		}
	}
}

func (m *Monitor) getRAMUtilization(ctx context.Context, interval time.Duration) error {
	type memInfo struct {
		Capacity uint64
	}

	var (
		memI            []memInfo
		capacity        uint64
		availableMemory float64
		highLoadCounter int
		avg             float64
		avgFormatted    string
	)

	query := "SELECT capacity FROM Win32_PhysicalMemory"
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

	query = "SELECT AvailableMBytes FROM Win32_PerfFormattedData_PerfOS_Memory"
	var memoryPoint []mem

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	buf := models.NewRingBuffer(5)

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
			buf.Add(availableMemory)

			avg = buf.GetAverage()
			if avg <= 25 {
				highLoadCounter++
			} else if highLoadCounter > 0 {
				highLoadCounter--
			}

			avgFormatted = fmt.Sprintf("%.*f", 2, avg)
			slog.Debug("", "available memory in percent", avgFormatted)

			m.ramUtilization.Lock()
			if avg <= 10 {
				m.ramUtilization.LoadZone = DangerZone
			} else if highLoadCounter >= 10 {
				m.ramUtilization.LoadZone = WarningZone
			} else {
				m.ramUtilization.LoadZone = NormalZone
			}
			m.ramUtilization.Value = avgFormatted
			m.ramUtilization.Unlock()
			memory.Set(avg)
		case <-ctx.Done():
			slog.Debug("RAM load monitoring is stopped")
			return nil
		}
	}
}

func (m *Monitor) getNetUtilization(ctx context.Context, interval time.Duration) error {
	var (
		netUtil         float64
		avg             float64
		highLoadCounter int
		netInfo         []net
		netName         []networkName
		avgFormatted    string
		err             error
	)

	query := "SELECT InterfaceDescription FROM MSFT_NetAdapter WHERE ConnectorPresent=1"
	err = wmi.QueryNamespace(query, &netName, `root\StandardCimv2`)
	if err != nil {
		return err
	}
	if len(netName) == 0 {
		return errors.New("no network data")
	}

	slog.Debug("", "network name", netName[0].InterfaceDescription)

	query = "SELECT CurrentBandwidth, BytesTotalPerSec FROM Win32_PerfFormattedData_Tcpip_NetworkInterface where Name = '" + netName[0].InterfaceDescription + "'"
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	buf := models.NewRingBuffer(5)

	for {
		select {
		case <-ticker.C:
			err = wmi.Query(query, &netInfo)
			if err != nil {
				return err
			}
			if len(netInfo) == 0 {
				return errors.New("no network data")
			}

			netUtil = 8 * float64(netInfo[0].BytesTotalPerSec) / float64(netInfo[0].CurrentBandwidth) * 100
			buf.Add(netUtil)

			avg = buf.GetAverage()
			if avg >= 80 {
				highLoadCounter++
			} else if highLoadCounter > 0 {
				highLoadCounter--
			}

			avgFormatted = fmt.Sprintf("%.*f", 2, avg)
			slog.Debug("", "network utilization", avgFormatted)
			m.netUtilization.Lock()
			if avg >= 90 {
				m.netUtilization.LoadZone = DangerZone
			} else if highLoadCounter >= 10 {
				m.netUtilization.LoadZone = WarningZone
			} else {
				m.netUtilization.LoadZone = NormalZone
			}
			m.netUtilization.Value = avgFormatted
			m.netUtilization.Unlock()
			network.Set(avg)
		case <-ctx.Done():
			slog.Debug("network load monitoring is stopped")
			return nil
		}
	}

}

func (m *Monitor) getDiskUtilization(ctx context.Context, interval time.Duration) error {
	var (
		diskInfo        []disk
		highLoadCounter int
		diskUtil        float64
		avg             float64
		avgFormatted    string
		err             error
	)
	query := "SELECT PercentDiskTime FROM Win32_PerfFormattedData_PerfDisk_PhysicalDisk WHERE Name = '_Total'"
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	buf := models.NewRingBuffer(5)
	for {
		select {
		case <-ticker.C:
			err = wmi.Query(query, &diskInfo)
			if err != nil {
				return err
			}
			if len(diskInfo) == 0 {
				return errors.New("no disk data")
			}

			diskUtil = float64(diskInfo[0].PercentDiskTime)
			buf.Add(diskUtil)

			avg = buf.GetAverage()
			if avg >= 80 {
				highLoadCounter++
			} else if highLoadCounter > 0 {
				highLoadCounter--
			}

			avgFormatted = fmt.Sprintf("%.*f", 2, avg)
			slog.Debug("", "disk utilization", avgFormatted)
			m.diskUtilization.Lock()
			if avg >= 90 {
				m.diskUtilization.LoadZone = DangerZone
			} else if highLoadCounter >= 10 {
				m.diskUtilization.LoadZone = WarningZone
			} else {
				m.diskUtilization.LoadZone = NormalZone
			}
			m.diskUtilization.Value = avgFormatted
			m.diskUtilization.Unlock()
			diskIO.Set(avg)
		case <-ctx.Done():
			slog.Debug("disk load monitoring is stopped")
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

func (m *Monitor) GetDiskUtilizationValue() *models.Utilization {
	return &m.diskUtilization
}
