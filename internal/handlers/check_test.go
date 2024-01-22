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

func TestCheck_CPUUtilizationNormalZone(t *testing.T) {
	m := &services.Monitor{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	c := configs.Checker{Interval: time.Millisecond}
	m.Start(ctx, c)

	time.Sleep(time.Millisecond * 5)

	router := NewRouter(m)

	q := m.GetCPUUtilizationValue()
	q.LoadZone = services.NormalZone
	req, _ := http.NewRequest("GET", "/check", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestCheck_CPUUtilizationWarningZone(t *testing.T) {
	m := &services.Monitor{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	c := configs.Checker{Interval: time.Microsecond}
	m.Start(ctx, c)

	time.Sleep(time.Millisecond * 5)

	router := NewRouter(m)

	q := m.GetCPUUtilizationValue()
	q.LoadZone = services.WarningZone
	req, _ := http.NewRequest("GET", "/check", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "CPU utilization exceeds 75%.", rr.Body.String())
}

func TestCheck_CPUUtilizationDangerZone(t *testing.T) {
	m := &services.Monitor{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	c := configs.Checker{Interval: time.Microsecond}
	m.Start(ctx, c)

	time.Sleep(time.Millisecond * 5)

	router := NewRouter(m)

	q := m.GetCPUUtilizationValue()
	q.LoadZone = services.DangerZone
	req, _ := http.NewRequest("GET", "/check", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusServiceUnavailable, rr.Code)
	assert.Equal(t, "CPU utilization exceeds 90%.", rr.Body.String())
}

func TestCheck_RAMUtilizationNormalZone(t *testing.T) {
	m := &services.Monitor{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	c := configs.Checker{Interval: time.Millisecond}
	m.Start(ctx, c)

	time.Sleep(time.Millisecond * 5)

	router := NewRouter(m)

	q := m.GetRAMUtilizationValue()
	q.LoadZone = services.NormalZone
	req, _ := http.NewRequest("GET", "/check", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestCheck_RAMUtilizationWarningZone(t *testing.T) {
	m := &services.Monitor{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	c := configs.Checker{Interval: time.Microsecond}
	m.Start(ctx, c)

	time.Sleep(time.Millisecond * 5)

	router := NewRouter(m)

	q := m.GetRAMUtilizationValue()
	q.LoadZone = services.WarningZone
	req, _ := http.NewRequest("GET", "/check", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "RAM utilization exceeds 75%.", rr.Body.String())
}

func TestCheck_RAMUtilizationDangerZone(t *testing.T) {
	m := &services.Monitor{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	c := configs.Checker{Interval: time.Microsecond}
	m.Start(ctx, c)

	time.Sleep(time.Millisecond * 5)

	router := NewRouter(m)

	q := m.GetRAMUtilizationValue()
	q.LoadZone = services.DangerZone
	req, _ := http.NewRequest("GET", "/check", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusServiceUnavailable, rr.Code)
	assert.Equal(t, "RAM utilization exceeds 90%.", rr.Body.String())
}

func TestCheck_NETUtilizationNormalZone(t *testing.T) {
	m := &services.Monitor{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	c := configs.Checker{Interval: time.Millisecond}
	m.Start(ctx, c)

	time.Sleep(time.Millisecond * 5)

	router := NewRouter(m)

	q := m.GetNetUtilizationValue()
	q.LoadZone = services.NormalZone
	req, _ := http.NewRequest("GET", "/check", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestCheck_NETUtilizationWarningZone(t *testing.T) {
	m := &services.Monitor{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	c := configs.Checker{Interval: time.Microsecond}
	m.Start(ctx, c)

	time.Sleep(time.Millisecond * 5)

	router := NewRouter(m)

	q := m.GetNetUtilizationValue()
	q.LoadZone = services.WarningZone
	req, _ := http.NewRequest("GET", "/check", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "Network utilization exceeds 75%.", rr.Body.String())
}

func TestCheck_NETUtilizationDangerZone(t *testing.T) {
	m := &services.Monitor{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	c := configs.Checker{Interval: time.Microsecond}
	m.Start(ctx, c)

	time.Sleep(time.Millisecond * 5)

	router := NewRouter(m)

	q := m.GetNetUtilizationValue()
	q.LoadZone = services.DangerZone
	req, _ := http.NewRequest("GET", "/check", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusServiceUnavailable, rr.Code)
	assert.Equal(t, "Network utilization exceeds 90%.", rr.Body.String())
}
