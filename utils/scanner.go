package utils

import (
	"bytes"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
)

type NoServiceEndpointsError struct {
	Message string
}

func (e *NoServiceEndpointsError) Error() string {
	return e.Message
}

var transport = &http.Transport{
	IdleConnTimeout: 30 * time.Second,
	MaxIdleConns:    10,
}
var insecureTransport = &http.Transport{
	IdleConnTimeout: 30 * time.Second,
	MaxIdleConns:    10,
	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
}
var httpClient = &http.Client{
	Timeout:   30 * time.Second,
	Transport: transport,
}
var insecureHTTPClient = &http.Client{
	Timeout:   30 * time.Second,
	Transport: insecureTransport,
}

// Scanner fetches the services in a config snapshot. It keeps no endpoint
// state: every GetServiceData call derives the endpoint list from the cfg it
// is given, so config changes apply on the next call with no Setup step,
// subscriber, or lock. The only state is the HTTP clients, which are fields
// so tests can inject their own (e.g. httptest.Server.Client()).
type Scanner struct {
	client   *http.Client
	insecure *http.Client
}

// NewScanner returns a Scanner using the shared pooled clients.
func NewScanner() *Scanner {
	return &Scanner{client: httpClient, insecure: insecureHTTPClient}
}

// clientFor returns the http.Client appropriate for the endpoint's
// SkipInsecure setting, so TLS verification can be relaxed per-service
// without duplicating the fetch/retry logic.
func (s *Scanner) clientFor(svc Service) *http.Client {
	if svc.SkipInsecure != nil && *svc.SkipInsecure {
		return s.insecure
	}
	return s.client
}

// Endpoints derives the endpoint list from a config snapshot, with Name
// filled from the map key and sorted by name so fetch/result order is stable.
// The returned slice is a fresh copy and safe to keep or modify.
func Endpoints(cfg *Config) ([]Service, error) {
	if cfg == nil {
		return nil, &NoConfigError{"configuration", "cannot be nil"}
	}
	if cfg.Services == nil {
		return nil, &NoServiceEndpointsError{"No service endpoints found in current configuration!"}
	}
	eps := make([]Service, 0, len(cfg.Services))
	for name, svc := range cfg.Services {
		svc.Name = name
		eps = append(eps, svc)
	}
	slices.SortFunc(eps, func(a, b Service) int { return strings.Compare(a.Name, b.Name) })
	return eps, nil
}

// GetServiceData fetches every service in cfg concurrently and returns the
// results. cfg is the caller's snapshot; it is used for both the endpoint
// list and Check, so one cycle never mixes two configs.
func (s *Scanner) GetServiceData(cfg *Config) (*ServiceResponse, error) {
	endpoints, err := Endpoints(cfg)
	if err != nil {
		return nil, err
	}

	results := make([]struct {
		data ServiceData
		tls  TlsStatus
	}, len(endpoints))
	var wg sync.WaitGroup

	slog.Info("Scanning services HTTP endpoints . . .")
	for i, ep := range endpoints {
		if !ep.IsActive() {
			slog.Info("Endpoint is disabled, continuing", "endpoint", ep.Name)
			continue
		}
		wg.Add(1)
		go func(i int, ep Service) {
			defer wg.Done()
			results[i].data, results[i].tls = s.fetchOne(ep)
		}(i, ep)
	}
	wg.Wait()

	var svcResponse ServiceResponse
	for _, r := range results {
		// Disabled endpoints are never fetched, leaving their slot as a
		// zero-value ServiceData{}; skip them so we don't insert blank
		// rows into the DB or have Check() flag them as down.
		if r.data.ServiceName == "" {
			continue
		}
		svcResponse.AllServices = append(svcResponse.AllServices, r.data)
		svcResponse.TlsData = append(svcResponse.TlsData, r.tls)
	}
	if svcResponse.DownServices, err = Check(cfg, svcResponse.AllServices); err != nil {
		return nil, err
	}
	return &svcResponse, nil
}

// fetchOne performs the HTTP fetch (plus optional API fetch) for a single endpoint.
func (s *Scanner) fetchOne(endpoint Service) (ServiceData, TlsStatus) {
	var sd ServiceData
	sd.Timestamp = time.Now()
	sd.ServiceName = endpoint.Name
	sd.ServiceURL = endpoint.URL
	sd.Active = true
	if endpoint.Description != nil {
		sd.ServiceDescription = *endpoint.Description
	}
	sd.SkipInsecure = endpoint.SkipInsecure != nil && *endpoint.SkipInsecure
	start := time.Now()

	res, err := s.do(endpoint)
	if err != nil {
		res, err = s.ErrorRetry(endpoint, err)
	}
	if err != nil {
		// no response means no handshake state to record
		slog.Error("Error fetching endpoint", "url", endpoint.URL, "error", err)
		sd.ServiceHTTPResponse = err.Error()
		sd.ServiceResponseTime = time.Since(start).Nanoseconds()
		return sd, TlsStatus{ServiceName: endpoint.Name}
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			slog.Error("Failed to close response body", "error", err)
		}
	}()

	tlsStatus, err := checkTls(endpoint.Name, res.TLS)
	if err != nil {
		slog.Error("failed to check TLS status", "service", endpoint.Name, "error", err)
	}

	sd.ServiceHTTPResponse = strconv.Itoa(res.StatusCode)
	if endpoint.API_URL != nil {
		if err := s.GetAPIData(endpoint, &sd, start); err != nil {
			slog.Error("Error getting API data", "endpoint", endpoint.Name, "error", err)
			sd.Error = true
			sd.ServiceAPIResponse = err.Error()
		}
	}
	// GetAPIData sets the response time when it runs; otherwise measure here
	if sd.ServiceResponseTime == 0 {
		sd.ServiceResponseTime = time.Since(start).Nanoseconds()
	}
	return sd, tlsStatus
}

// do sends one GET to the endpoint's URL with the GoUp User-Agent.
func (s *Scanner) do(endpoint Service) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, endpoint.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("User-Agent", "GoUp/"+Version)
	return s.clientFor(endpoint).Do(req)
}

// ErrorRetry retries the fetch Retry_Requests times. With no retries
// configured it returns initialErr unchanged.
func (s *Scanner) ErrorRetry(endpoint Service, initialErr error) (*http.Response, error) {
	currentErr := initialErr
	if endpoint.Retry_Requests != nil {
		for range *endpoint.Retry_Requests {
			slog.Warn("Error fetching endpoint, re-attempting GET request", "url", endpoint.URL, "error", currentErr)
			res, err := s.do(endpoint)
			if err == nil {
				return res, nil
			}
			currentErr = err
		}
	}
	return nil, currentErr
}

// GetAPIData fetches the endpoint's API_URL into sd.ServiceAPIResponse.
func (s *Scanner) GetAPIData(endpoint Service, sd *ServiceData, start time.Time) error {
	apiReq, err := http.NewRequest(http.MethodGet, *endpoint.API_URL, nil)
	if err != nil {
		return err
	}
	if endpoint.API_KEY != nil {
		apiReq.Header.Set("Authorization", "Bearer "+*endpoint.API_KEY)
	}
	apiReq.Header.Set("Content-Type", "application/json")
	apiReq.Header.Set("User-Agent", "GoUp/"+Version)

	apiRes, err := s.clientFor(endpoint).Do(apiReq)
	if err != nil {
		slog.Error("Error occured during API request", "error", err)
		return err
	}
	defer func() {
		if err := apiRes.Body.Close(); err != nil {
			slog.Error("Error closing API response body", "error", err)
		}
	}()

	apiBody, err := io.ReadAll(apiRes.Body)
	if err != nil {
		return err
	}
	sd.ServiceAPIResponse = string(apiBody)
	sd.ServiceResponseTime = time.Since(start).Nanoseconds()
	slog.Info("Fetched API data successfully")
	return nil
}

func chainExpiry(state *tls.ConnectionState) (soonest time.Time, cert *x509.Certificate, ok bool) {
	if state == nil || len(state.PeerCertificates) == 0 {
		return time.Time{}, nil, false // plain HTTP, or handshake gave us nothing
	}
	for _, c := range state.PeerCertificates {
		// a server can include a self-signed root in what it sends;
		// that root's expiry isn't the operator's concern, so skip it
		if bytes.Equal(c.RawSubject, c.RawIssuer) {
			continue
		}
		if cert == nil || c.NotAfter.Before(cert.NotAfter) {
			cert = c
		}
	}
	if cert == nil { // chain was only a self-signed cert (internal CA, etc.)
		cert = state.PeerCertificates[0]
	}
	return cert.NotAfter, cert, true
}

func checkTls(svcName string, state *tls.ConnectionState) (TlsStatus, error) {
	status := TlsStatus{ServiceName: svcName}

	soonest, c, ok := chainExpiry(state)
	if !ok {
		slog.Info("plain HTTP or empty chain", "service", svcName)
		return status, nil // plain HTTP or empty chain — nothing to record
	}
	sum := sha256.Sum256(c.Raw)
	now := time.Now()

	status.Fingerprint = hex.EncodeToString(sum[:])
	status.Not_after = soonest
	status.Subject = c.Subject.CommonName
	status.Issuer = c.Issuer.CommonName
	status.Is_expired = isExpired(c.NotAfter) // a field, not the gate
	var chain []CertJSON
	for _, cc := range state.PeerCertificates {
		chain = append(chain, CertJSON{cc.Subject.CommonName, cc.Issuer.CommonName, cc.NotAfter, cc.IsCA})
	}
	b, err := json.Marshal(chain)
	if err != nil {
		return status, fmt.Errorf("failed to marshal certificate chain: %w", err)
	}
	status.Chain = string(b)
	status.First_seen = now // upsert CASE preserves real value
	status.Last_checked = now
	return status, nil
}
