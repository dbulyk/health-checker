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

func TestCheck_CPUUtilisationNormalZone(t *testing.T) {
	m := &services.Monitor{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	c := configs.Checker{Interval: time.Millisecond}
	m.Start(ctx, c)

	time.Sleep(time.Millisecond * 5)

	router := NewRouter(m)

	q := m.GetCPUUtilisationValue()
	q.LoadZone = services.NormalZone
	req, _ := http.NewRequest("GET", "/check", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestCheck_CPUUtilisationWarningZone(t *testing.T) {
	m := &services.Monitor{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	c := configs.Checker{Interval: time.Microsecond}
	m.Start(ctx, c)

	time.Sleep(time.Millisecond * 5)

	router := NewRouter(m)

	q := m.GetCPUUtilisationValue()
	q.LoadZone = services.WarningZone
	req, _ := http.NewRequest("GET", "/check", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "CPU utilisation exceeds 75%.", rr.Body.String())
}

func TestCheck_CPUUtilisationDangerZone(t *testing.T) {
	m := &services.Monitor{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	c := configs.Checker{Interval: time.Microsecond}
	m.Start(ctx, c)

	time.Sleep(time.Millisecond * 5)

	router := NewRouter(m)

	q := m.GetCPUUtilisationValue()
	q.LoadZone = services.DangerZone
	req, _ := http.NewRequest("GET", "/check", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusServiceUnavailable, rr.Code)
}

func TestCheck_RAMUtilisationNormalZone(t *testing.T) {
	m := &services.Monitor{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	c := configs.Checker{Interval: time.Millisecond}
	m.Start(ctx, c)

	time.Sleep(time.Millisecond * 5)

	router := NewRouter(m)

	q := m.GetRAMUtilisationValue()
	q.LoadZone = services.NormalZone
	req, _ := http.NewRequest("GET", "/check", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestCheck_RAMUtilisationWarningZone(t *testing.T) {
	m := &services.Monitor{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	c := configs.Checker{Interval: time.Microsecond}
	m.Start(ctx, c)

	time.Sleep(time.Millisecond * 5)

	router := NewRouter(m)

	q := m.GetRAMUtilisationValue()
	q.LoadZone = services.WarningZone
	req, _ := http.NewRequest("GET", "/check", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "RAM utilisation exceeds 75%.", rr.Body.String())
}

func TestCheck_RAMUtilisationDangerZone(t *testing.T) {
	m := &services.Monitor{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	c := configs.Checker{Interval: time.Microsecond}
	m.Start(ctx, c)

	time.Sleep(time.Millisecond * 5)

	router := NewRouter(m)

	q := m.GetRAMUtilisationValue()
	q.LoadZone = services.DangerZone
	req, _ := http.NewRequest("GET", "/check", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusServiceUnavailable, rr.Code)
}
