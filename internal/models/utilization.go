package models

import "sync"

type Utilization struct {
	sync.Mutex
	Percentages float64
}

type Win32PerfFormattedDataPerfOsProcessor struct {
	PercentProcessorTime uint64
	TimeStamp_Sys100NS   uint64
}

type Win32PerfFormattedDataPerfOsMemory struct {
	PercentCommittedBytesInUse uint64
}
