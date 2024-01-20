package models

import "sync"

type Utilization struct {
	sync.Mutex
	Value    float64
	HighLoad bool
}

type Processor struct {
	PercentProcessorTime uint64
	TimeStamp_Sys100NS   uint64
}

//type Win32PerfFormattedDataPerfOsMemory struct {
//	PercentCommittedBytesInUse uint64
//}
