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

func TestCheck_UtilisationNormalZone(t *testing.T) {
	m := &services.Monitor{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	c := configs.Checker{Interval: time.Millisecond}
	m.Start(ctx, c)

	time.Sleep(time.Millisecond * 5)

	router := NewRouter(m)

	q := m.GetCPUUtilisationValue()
	q.LoadZone = "normal"
	req, _ := http.NewRequest("GET", "/check", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestCheck_UtilisationWarningZone(t *testing.T) {
	m := &services.Monitor{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	c := configs.Checker{Interval: time.Microsecond}
	m.Start(ctx, c)

	time.Sleep(time.Millisecond * 5)

	router := NewRouter(m)

	q := m.GetCPUUtilisationValue()
	q.LoadZone = "warning"
	req, _ := http.NewRequest("GET", "/check", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "CPU utilisation exceeds 75%.", rr.Body.String())
}

func TestCheck_UtilisationDangerZone(t *testing.T) {
	m := &services.Monitor{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	c := configs.Checker{Interval: time.Microsecond}
	m.Start(ctx, c)

	time.Sleep(time.Millisecond * 5)

	router := NewRouter(m)

	q := m.GetCPUUtilisationValue()
	q.LoadZone = "danger"
	req, _ := http.NewRequest("GET", "/check", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusServiceUnavailable, rr.Code)
}
