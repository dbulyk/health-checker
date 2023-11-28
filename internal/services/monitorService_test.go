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
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	monitor.Start(ctx, configs.Checker{Interval: time.Second})

	time.Sleep(time.Second * 5)

	assert.NotZero(t, monitor.GetCPUUtilizationValue())
	assert.NotZero(t, monitor.GetRAMUtilizationValue())
}

func TestMonitor_GetCPUUtilization(t *testing.T) {
	monitor := NewMonitor()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	err := monitor.GetCPUUtilization(ctx, time.Second)

	assert.Equal(t, err, context.DeadlineExceeded)
	assert.NotZero(t, monitor.GetCPUUtilizationValue())
}

func TestMonitor_GetRAMUtilization(t *testing.T) {
	monitor := NewMonitor()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	err := monitor.GetRAMUtilization(ctx, time.Second)

	assert.Nil(t, err)
	assert.NotZero(t, monitor.GetRAMUtilizationValue())
}
