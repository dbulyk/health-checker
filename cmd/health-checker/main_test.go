package main

import (
	"health-checker/config"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUpdateCPULoad(t *testing.T) {
	interval := 1 * time.Second
	sigs := make(chan os.Signal)

	go func() {
		err := updateCPULoad(interval, sigs)
		assert.NoError(t, err)
	}()

	time.Sleep(3 * interval)

	sigs <- os.Interrupt
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
