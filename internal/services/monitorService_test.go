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

	assert.NotNil(t, monitor.GetCPUUtilizationValue())
	//assert.NotZero(t, monitor.GetRAMUtilizationValue())
}

func TestMonitor_GetCPUUtilization(t *testing.T) {
	monitor := NewMonitor()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err := monitor.GetCPUUtilization(ctx, time.Millisecond)

	assert.NoError(t, err)
	assert.NotZero(t, monitor.GetCPUUtilizationValue())
}

//func TestMonitor_GetRAMUtilization(t *testing.T) {
//	monitor := NewMonitor()
//	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
//	defer cancel()
//
//	err := monitor.GetRAMUtilization(ctx, time.Second)
//
//	assert.Nil(t, err)
//	assert.NotZero(t, monitor.GetRAMUtilizationValue())
//}
