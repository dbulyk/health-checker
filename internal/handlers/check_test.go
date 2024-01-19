package handlers

//
//import (
//	"context"
//	"health-checker/internal/configs"
//	"health-checker/internal/services"
//	"net/http"
//	"net/http/httptest"
//	"testing"
//	"time"
//
//	"github.com/stretchr/testify/assert"
//)
//
//func TestCheck_UtilizationUnderThreshold(t *testing.T) {
//	m := &services.Monitor{}
//	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
//	defer cancel()
//
//	c := configs.Checker{Interval: time.Second, Threshold: 80.0}
//	m.Start(ctx, c)
//
//	time.Sleep(time.Second * 5)
//
//	router := NewRouter(m, c)
//
//	req, _ := http.NewRequest("GET", "/check", nil)
//	rr := httptest.NewRecorder()
//
//	router.ServeHTTP(rr, req)
//
//	assert.Equal(t, http.StatusOK, rr.Code)
//	assert.Contains(t, rr.Body.String(), "Утилизация процессора")
//	assert.Contains(t, rr.Body.String(), "Утилизация памяти")
//}
//
//func TestCheck_UtilizationOverThreshold(t *testing.T) {
//	m := &services.Monitor{}
//	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
//	defer cancel()
//
//	c := configs.Checker{Interval: time.Second, Threshold: 0.1}
//	m.Start(ctx, c)
//
//	time.Sleep(time.Second * 3)
//
//	router := NewRouter(m, c)
//
//	req, _ := http.NewRequest("GET", "/check", nil)
//	rr := httptest.NewRecorder()
//
//	router.ServeHTTP(rr, req)
//
//	assert.Equal(t, http.StatusServiceUnavailable, rr.Code)
//	assert.Contains(t, rr.Body.String(), "Утилизация процессора")
//	assert.Contains(t, rr.Body.String(), "Утилизация памяти")
//}
