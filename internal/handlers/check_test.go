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

func TestCheck_UtilizationUnderThreshold(t *testing.T) {
	m := &services.Monitor{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	c := configs.Checker{Interval: time.Millisecond}
	m.Start(ctx, c)

	time.Sleep(time.Millisecond * 5)

	router := NewRouter(m, c)

	req, _ := http.NewRequest("GET", "/check", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

//TODO: fix this test using mock
//func TestCheck_UtilizationOverThreshold(t *testing.T) {
//	m := &services.Monitor{}
//	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
//	defer cancel()
//
//	c := configs.Checker{Interval: time.Microsecond}
//	m.Start(ctx, c)
//
//	time.Sleep(time.Millisecond * 5)
//
//	router := NewRouter(m, c)
//
//	req, _ := http.NewRequest("GET", "/check", nil)
//	rr := httptest.NewRecorder()
//
//	router.ServeHTTP(rr, req)
//
//	assert.Equal(t, http.StatusServiceUnavailable, rr.Code)
//}
