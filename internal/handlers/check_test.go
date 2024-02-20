package handlers

import (
	"context"
	"health-checker/internal/configs"
	"health-checker/internal/services"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_CheckUtilization_AllNormal(t *testing.T) {
	m := &services.Monitor{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()

	c := configs.Checker{
		Interval: time.Microsecond,
	}
	m.Start(ctx, c)

	time.Sleep(time.Millisecond * 5)

	router := NewRouter(m)

	req, _ := http.NewRequest("GET", "/check", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "CPU")
	assert.Contains(t, rr.Body.String(), "RAM")
	assert.Contains(t, rr.Body.String(), "Network")
	assert.Contains(t, rr.Body.String(), "Disk")

	ctx.Done()
	time.Sleep(time.Second)
}

func Test_CheckUtilization_WithWarningZone(t *testing.T) {
	m := &services.Monitor{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	c := configs.Checker{
		Interval: time.Microsecond,
	}
	m.Start(ctx, c)

	time.Sleep(time.Millisecond * 5)

	router := NewRouter(m)

	q := m.GetDiskUtilizationValue()
	q.LoadZone = services.WarningZone
	req, _ := http.NewRequest("GET", "/check", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "CPU")
	assert.Contains(t, rr.Body.String(), "RAM")
	assert.Contains(t, rr.Body.String(), "Network")
	assert.Contains(t, rr.Body.String(), "Disk")
	assert.Contains(t, rr.Body.String(), "Warning: High utilization")

	ctx.Done()
	time.Sleep(time.Second)
}

func Test_CheckUtilization_WithDangerZone(t *testing.T) {
	m := &services.Monitor{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	c := configs.Checker{
		Interval: time.Microsecond,
	}
	m.Start(ctx, c)

	time.Sleep(time.Millisecond * 5)

	router := NewRouter(m)

	q := m.GetDiskUtilizationValue()
	q.LoadZone = services.DangerZone
	req, _ := http.NewRequest("GET", "/check", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusServiceUnavailable, rr.Code)
	assert.Contains(t, rr.Body.String(), "CPU")
	assert.Contains(t, rr.Body.String(), "RAM")
	assert.Contains(t, rr.Body.String(), "Network")
	assert.Contains(t, rr.Body.String(), "Disk")
	assert.Contains(t, rr.Body.String(), "Danger: Critical utilization")

	ctx.Done()
	time.Sleep(time.Second)
}
