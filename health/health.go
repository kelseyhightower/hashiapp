package health

import (
	"net/http"
	"sync"
)

var (
	healthzStatus = http.StatusOK
	mu            sync.RWMutex
)

func HealthzStatus() int {
	mu.RLock()
	defer mu.RUnlock()
	return healthzStatus
}

func SetHealthzStatus(status int) {
	mu.Lock()
	healthzStatus = status
	mu.Unlock()
}

// HealthzHandler responds to health check requests.
func HealthzHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(HealthzStatus())
}

func HealthzStatusHandler(w http.ResponseWriter, r *http.Request) {
	switch HealthzStatus() {
	case http.StatusOK:
		SetHealthzStatus(http.StatusServiceUnavailable)
	case http.StatusServiceUnavailable:
		SetHealthzStatus(http.StatusOK)
	}
	w.WriteHeader(http.StatusOK)
}
