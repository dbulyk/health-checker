package services

import (
	"context"
	"health-checker/internal/configs"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMonitor_Start(t *testing.T) {
	monitor := NewMonitor()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	monitor.Start(ctx, configs.Checker{Interval: time.Microsecond})

	time.Sleep(time.Millisecond * 5)

	assert.NotNil(t, monitor.GetCPUUtilisationValue())
	assert.NotZero(t, monitor.GetRAMUtilisationValue())
}
