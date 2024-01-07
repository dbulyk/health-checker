package services

import (
	"context"
	"errors"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/yusufpapurcu/wmi"
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

var (
	startPoint []models.Win32PerfFormattedDataPerfOsProcessor
	endPoint   []models.Win32PerfFormattedDataPerfOsProcessor
)

func NewMonitor() *Monitor {
	return &Monitor{}
}

func (m *Monitor) Start(ctx context.Context, cfg configs.Checker) {
	go func() {
		slog.Debug("начат мониторинг загрузки процессора")

		err := m.GetCPUUtilization(ctx, cfg.Interval)
		if err != nil {
			slog.Error("ошибка получения данных процессора", "ошибка", err)
		}
	}()

	//go func() {
	//	slog.Debug("ram load started")
	//
	//	err := m.GetRAMUtilization(ctx, cfg.Interval)
	//	if err != nil {
	//		slog.Error("ram load error", "error", err)
	//	}
	//}()
}

func (m *Monitor) GetCPUUtilization(ctx context.Context, interval time.Duration) error {
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
				return errors.New("нет данных о процессоре")
			}

			startPointProcTime := startPoint[0].PercentProcessorTime
			startPointTS := startPoint[0].TimeStamp_Sys100NS

			time.Sleep(interval)

			err = wmi.Query(query, &endPoint)
			if err != nil {
				return err
			}

			if len(endPoint) == 0 {
				return errors.New("нет данных о процессоре")
			}

			endPointProcTime := endPoint[0].PercentProcessorTime
			endPointTS := endPoint[0].TimeStamp_Sys100NS

			/*
				механизм расчета загрузки процессора
				основан на https://learn.microsoft.com/en-us/windows/win32/wmisdk/monitoring-performance-data#using-raw-performance-data-classes
			*/
			cpuUtil := (1.0 - float64(endPointProcTime-startPointProcTime)/float64(endPointTS-startPointTS)) * 100

			m.cpuUtilization.Lock()
			m.cpuUtilization.Percentages = cpuUtil
			slog.Debug("", "Загрузка процессора", cpuUtil)
			m.cpuUtilization.Unlock()

		case <-ctx.Done():
			slog.Debug("мониторинг загрузки процессора остановлен")
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
