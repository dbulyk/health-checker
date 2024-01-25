package models

import "sync"

type Utilization struct {
	sync.Mutex
	Value    string
	LoadZone string
}
