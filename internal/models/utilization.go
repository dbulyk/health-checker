package models

import "sync"

type Utilization struct {
	sync.Mutex
	Value    float64
	LoadZone string
}
