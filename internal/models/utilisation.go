package models

import "sync"

type Utilisation struct {
	sync.Mutex
	Value    float64
	LoadZone string
}

type Processor struct {
	PercentProcessorTime uint64
	TimeStamp_Sys100NS   uint64
}

//type Win32PerfFormattedDataPerfOsMemory struct {
//	PercentCommittedBytesInUse uint64
//}
