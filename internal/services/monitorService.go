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

type Monitor struct {
	cpuUtilisation models.Utilisation
	//ramutilisation models.utilisation
	PollCount atomic.Int64
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

	//go func() {
	//	slog.Debug("начат мониторинг загрузки памяти")
	//
	//	err := m.GetRAMutilisation(ctx, cfg.Interval)
	//	if err != nil {
	//		slog.Error("ошибка получения данных памяти", "error", err)
	//	}
	//}()
}

func (m *Monitor) GetCPUUtilisation(ctx context.Context, interval time.Duration) error {
	var (
		startPoint         []models.Processor
		endPoint           []models.Processor
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

//func (m *Monitor) GetRAMutilisation(ctx context.Context, interval time.Duration) error {
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
//			m.ramutilisation.Lock()
//			if m.PollCount.Load() > limitPollCount {
//				m.ramutilisation.Value = memoryUsage
//				m.PollCount.Swap(1)
//			} else {
//				m.ramutilisation.Value += memoryUsage
//				m.PollCount.Add(1)
//			}
//			slog.Debug("", "Загрузка памяти", m.ramutilisation.Value/float64(m.PollCount.Load()))
//			m.ramutilisation.Unlock()
//		case <-ctx.Done():
//			slog.Debug("мониторинг загрузки памяти остановлен")
//			return nil
//		}
//	}
//}

func (m *Monitor) GetCPUUtilisationValue() *models.Utilisation {
	m.cpuUtilisation.Lock()
	defer m.cpuUtilisation.Unlock()

	return &m.cpuUtilisation
}

//func (m *Monitor) GetRAMutilisationValue() float64 {
//	m.ramutilisation.Lock()
//	defer m.ramutilisation.Unlock()
//
//	return m.ramutilisation.Value / float64(m.PollCount.Load())
//}
