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

type Monitor struct {
	cpuUtilization models.Utilization
	//ramUtilization models.Utilization
	PollCount atomic.Int64
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

	//go func() {
	//	slog.Debug("начат мониторинг загрузки памяти")
	//
	//	err := m.GetRAMUtilization(ctx, cfg.Interval)
	//	if err != nil {
	//		slog.Error("ошибка получения данных памяти", "error", err)
	//	}
	//}()
}

func (m *Monitor) GetCPUUtilization(ctx context.Context, interval time.Duration) error {
	var (
		startPoint []models.Processor
		endPoint   []models.Processor
	)

	const query = "SELECT * FROM Win32_PerfRawData_PerfOS_Processor WHERE Name = '_Total'"

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	highLoadCounter := 0

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

			startPointProcTime := startPoint[0].PercentProcessorTime
			startPointTS := startPoint[0].TimeStamp_Sys100NS

			time.Sleep(interval)

			err = wmi.Query(query, &endPoint)
			if err != nil {
				return err
			}

			if len(endPoint) == 0 {
				return errors.New("no processor data")
			}

			endPointProcTime := endPoint[0].PercentProcessorTime
			endPointTS := endPoint[0].TimeStamp_Sys100NS

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
			m.cpuUtilization.Percentages = cpuUtil
			if highLoadCounter > 10 {
				m.cpuUtilization.HighLoad = true
			} else {
				m.cpuUtilization.HighLoad = false
			}

			slog.Debug("", "CPU load", cpuUtil)
			m.cpuUtilization.Unlock()
		case <-ctx.Done():
			slog.Debug("CPU load monitoring is stopped")
			return nil
		}
	}
}

//func (m *Monitor) GetRAMUtilization(ctx context.Context, interval time.Duration) error {
//	const query = "SELECT * FROM Win32_PerfFormattedData_PerfOS_Memory"
//	var memoryPoint []models.Win32PerfFormattedDataPerfOsMemory
//
//	ticker := time.NewTicker(interval)
//	defer ticker.Stop()
//
//	for {
//		select {
//		case <-ticker.C:
//			err := wmi.Query(query, &memoryPoint)
//			if err != nil {
//				log.Fatal(err)
//			}
//
//			if len(memoryPoint) == 0 {
//				return errors.New("нет данных о памяти")
//			}
//
//			memoryUsage := float64(memoryPoint[0].PercentCommittedBytesInUse)
//			limitPollCount := int64(5)
//
//			m.ramUtilization.Lock()
//			if m.PollCount.Load() > limitPollCount {
//				m.ramUtilization.Percentages = memoryUsage
//				m.PollCount.Swap(1)
//			} else {
//				m.ramUtilization.Percentages += memoryUsage
//				m.PollCount.Add(1)
//			}
//			slog.Debug("", "Загрузка памяти", m.ramUtilization.Percentages/float64(m.PollCount.Load()))
//			m.ramUtilization.Unlock()
//		case <-ctx.Done():
//			slog.Debug("мониторинг загрузки памяти остановлен")
//			return nil
//		}
//	}
//}

func (m *Monitor) GetCPUUtilizationValue() float64 {
	m.cpuUtilization.Lock()
	defer m.cpuUtilization.Unlock()

	return m.cpuUtilization.Percentages
}

//func (m *Monitor) GetRAMUtilizationValue() float64 {
//	m.ramUtilization.Lock()
//	defer m.ramUtilization.Unlock()
//
//	return m.ramUtilization.Percentages / float64(m.PollCount.Load())
//}
