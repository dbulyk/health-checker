package services

import (
	"context"
	"health-checker/internal/configs"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_Monitor_Start(t *testing.T) {
	monitor := NewMonitor()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	monitor.Start(ctx, configs.Checker{Interval: time.Microsecond})

	time.Sleep(time.Millisecond * 5)

	assert.NotNil(t, monitor.GetCPUUtilizationValue())
	assert.NotZero(t, monitor.GetCPUUtilizationValue())

	assert.NotNil(t, monitor.GetRAMUtilizationValue())
	assert.NotZero(t, monitor.GetRAMUtilizationValue())

	assert.NotNil(t, monitor.GetNetUtilizationValue())
	assert.NotZero(t, monitor.GetNetUtilizationValue())

	assert.NotNil(t, monitor.GetDiskUtilizationValue())
	assert.NotZero(t, monitor.GetDiskUtilizationValue())
}
