package models

import "sync"

type Utilisation struct {
	sync.Mutex
	Value    float64
	LoadZone string
}
