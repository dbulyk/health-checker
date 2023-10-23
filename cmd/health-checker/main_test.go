package main

import (
	"context"
	"health-checker/config"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUpdateCPULoad(t *testing.T) {
	interval := 2 * time.Second
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		time.Sleep(5 * time.Second)
		cancel()
	}()

	err := updateCPULoad(ctx, interval)
	assert.NoError(t, err)
}

func TestCheckCPUAndRAMLoad(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	assert.NoError(t, err)
	w := httptest.NewRecorder()

	loadLock.Lock()
	lastCPULoad = 60.0
	loadLock.Unlock()

	cfg = config.GetCheckerCfg()

	checkCPUAndRAMLoad(w, req)

	expectedStatus := http.StatusOK
	assert.Equal(t, expectedStatus, w.Code, "status code %d was expected, but received %d", expectedStatus, w.Code)

	loadLock.Lock()
	lastCPULoad = 90.0
	loadLock.Unlock()

	w = httptest.NewRecorder()
	checkCPUAndRAMLoad(w, req)

	expectedStatus = http.StatusServiceUnavailable
	assert.Equal(t, expectedStatus, w.Code, "status code %d was expected, but received %d", expectedStatus, w.Code)

	loadLock.Lock()
	lastCPULoad = 0.0
	loadLock.Unlock()
}
