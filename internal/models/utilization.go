package models

import "sync"

type Utilization struct {
	Percentages float64
	sync.Mutex
}
