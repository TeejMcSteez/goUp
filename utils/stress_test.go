package utils_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"goUp/utils"
)

// TestStress verifies correctness and exposes concurrency characteristics when
// checking a large number of services. A simulated per-request latency is
// applied so that the concurrency factor metric is meaningful:
//
//	~1.00 → fully sequential   (total svc time ≈ wall time)
//	~N    → fully parallel     (wall time ≈ single request latency)
func TestStress(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	const (
		numServices    = 100
		simulatedDelay = 25 * time.Millisecond
		maxTotalTime   = 60 * time.Second
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(simulatedDelay)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	services := make([]utils.Service, numServices)
	cfgServices := make(map[string]utils.Service, numServices)
	for i := range numServices {
		name := fmt.Sprintf("stress-service-%03d", i+1)
		svc := utils.Service{Name: name, URL: server.URL}
		services[i] = svc
		cfgServices[name] = svc
	}

	utils.SetServiceEndpoints(services)
	utils.Current_Config = &utils.Config{Services: cfgServices}
	defer func() {
		utils.Current_Config = nil
		utils.SetServiceEndpoints([]utils.Service{})
	}()

	start := time.Now()
	resp, err := utils.GetServiceData()
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("GetServiceData() failed: %v", err)
	}
	if len(resp.AllServices) != numServices {
		t.Errorf("expected %d results, got %d", numServices, len(resp.AllServices))
	}
	if len(resp.DownServices) != 0 {
		t.Errorf("expected 0 down services, got %d", len(resp.DownServices))
	}

	var totalSvcTime time.Duration
	var minSvcTime = time.Duration(1<<63 - 1)
	var maxSvcTime time.Duration
	for _, sd := range resp.AllServices {
		d, parseErr := time.ParseDuration(sd.ServiceResponseTime)
		if parseErr != nil {
			continue
		}
		totalSvcTime += d
		if d < minSvcTime {
			minSvcTime = d
		}
		if d > maxSvcTime {
			maxSvcTime = d
		}
	}

	t.Logf("Stress results (%d services, %s simulated latency per request):", numServices, simulatedDelay)
	t.Logf("  Wall time total:    %s", elapsed.Round(time.Millisecond))
	t.Logf("  Avg per service:    %.1fms", float64(elapsed.Milliseconds())/float64(numServices))
	t.Logf("  Min response time:  %s", minSvcTime.Round(time.Millisecond))
	t.Logf("  Max response time:  %s", maxSvcTime.Round(time.Millisecond))
	t.Logf("  Concurrency factor: %.2fx  (1.00 = fully sequential, %.0f.00 = fully parallel)",
		float64(totalSvcTime)/float64(elapsed), float64(numServices))

	if elapsed > maxTotalTime {
		t.Errorf("stress test took %s, exceeding %s safety limit", elapsed.Round(time.Millisecond), maxTotalTime)
	}
}
