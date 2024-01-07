package models

import "sync"

type Utilization struct {
	Percentages float64
	sync.Mutex
}

type Win32PerfFormattedDataPerfOsProcessor struct {
	PercentProcessorTime uint64
	TimeStamp_Sys100NS   uint64
}

type Win32PerfFormattedDataPerfOsMemory struct {
	PercentCommittedBytesInUse uint64
}
