package utils

import (
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestIsExpired(t *testing.T) {
	if !isExpired(time.Now().Add(-time.Hour)) {
		t.Error("Expected past time to be expired")
	}
	if isExpired(time.Now().Add(time.Hour)) {
		t.Error("Expected future time to not be expired")
	}
}

func TestChainExpiryNilState(t *testing.T) {
	soonest, cert, ok := chainExpiry(nil)
	if ok {
		t.Error("Expected ok=false for nil state")
	}
	if cert != nil {
		t.Error("Expected nil cert for nil state")
	}
	if !soonest.IsZero() {
		t.Error("Expected zero time for nil state")
	}
}

func TestChainExpiryEmptyCertificates(t *testing.T) {
	state := &tls.ConnectionState{PeerCertificates: []*x509.Certificate{}}
	_, cert, ok := chainExpiry(state)
	if ok {
		t.Error("Expected ok=false for empty peer certificates")
	}
	if cert != nil {
		t.Error("Expected nil cert for empty peer certificates")
	}
}

func TestChainExpirySkipsSelfSignedRoot(t *testing.T) {
	rootSubject := []byte("root-subject")
	leafExpiry := time.Now().Add(30 * 24 * time.Hour)
	rootExpiry := time.Now().Add(365 * 24 * time.Hour)

	root := &x509.Certificate{
		RawSubject: rootSubject,
		RawIssuer:  rootSubject, // self-signed
		NotAfter:   rootExpiry,
	}
	leaf := &x509.Certificate{
		RawSubject: []byte("leaf-subject"),
		RawIssuer:  []byte("intermediate-issuer"),
		NotAfter:   leafExpiry,
	}

	state := &tls.ConnectionState{PeerCertificates: []*x509.Certificate{root, leaf}}
	soonest, cert, ok := chainExpiry(state)
	if !ok {
		t.Fatal("Expected ok=true")
	}
	if cert != leaf {
		t.Errorf("Expected leaf certificate to be picked, got %+v", cert)
	}
	if !soonest.Equal(leafExpiry) {
		t.Errorf("Expected soonest expiry %v, got %v", leafExpiry, soonest)
	}
}

func TestChainExpiryPicksSoonestExpiring(t *testing.T) {
	now := time.Now()
	certA := &x509.Certificate{
		RawSubject: []byte("a-subject"),
		RawIssuer:  []byte("issuer"),
		NotAfter:   now.Add(60 * 24 * time.Hour),
	}
	certB := &x509.Certificate{
		RawSubject: []byte("b-subject"),
		RawIssuer:  []byte("issuer"),
		NotAfter:   now.Add(10 * 24 * time.Hour), // soonest
	}

	state := &tls.ConnectionState{PeerCertificates: []*x509.Certificate{certA, certB}}
	soonest, cert, ok := chainExpiry(state)
	if !ok {
		t.Fatal("Expected ok=true")
	}
	if cert != certB {
		t.Error("Expected the soonest-expiring cert to be picked")
	}
	if !soonest.Equal(certB.NotAfter) {
		t.Errorf("Expected soonest %v, got %v", certB.NotAfter, soonest)
	}
}

func TestChainExpiryFallsBackToSelfSignedOnly(t *testing.T) {
	subject := []byte("self-signed-subject")
	expiry := time.Now().Add(90 * 24 * time.Hour)
	root := &x509.Certificate{
		RawSubject: subject,
		RawIssuer:  subject,
		NotAfter:   expiry,
	}

	state := &tls.ConnectionState{PeerCertificates: []*x509.Certificate{root}}
	soonest, cert, ok := chainExpiry(state)
	if !ok {
		t.Fatal("Expected ok=true when the chain is only a self-signed cert")
	}
	if cert != root {
		t.Error("Expected fallback to the self-signed cert")
	}
	if !soonest.Equal(expiry) {
		t.Errorf("Expected soonest %v, got %v", expiry, soonest)
	}
}

// TestFetchOneHandlesRequestErrorWithoutRetries is a regression test for a
// nil pointer dereference: when the initial request fails and the endpoint
// has no Retry_Requests configured, ErrorRetry returns a nil *http.Response,
// but fetchOne used to dereference res.TLS unconditionally in that path.
func TestFetchOneHandlesRequestErrorWithoutRetries(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	url := server.URL
	server.Close() // requests to url now fail immediately with no response

	svc := Service{Name: "unreachable-service", URL: url}

	sd, status := fetchOne(svc)

	if sd.ServiceName != "unreachable-service" {
		t.Errorf("Expected ServiceName to be set, got %q", sd.ServiceName)
	}
	if sd.ServiceHTTPResponse == "" {
		t.Error("Expected ServiceHTTPResponse to contain the request error, got empty string")
	}
	if status.ServiceName != "unreachable-service" {
		t.Errorf("Expected TlsStatus.ServiceName to be set even without a response, got %q", status.ServiceName)
	}
}

func TestCheckTlsPlainHTTP(t *testing.T) {
	status, err := checkTls("plain-service", nil)
	if err == nil {
		t.Fatal("Expected error for nil TLS state")
	}
	if status.ServiceName != "plain-service" {
		t.Errorf("Expected ServiceName to be set even on error, got %q", status.ServiceName)
	}
}

func TestCheckTlsPopulatesStatus(t *testing.T) {
	rawCert := []byte("fake-der-bytes-for-testing")
	notAfter := time.Now().Add(-24 * time.Hour) // expired
	cert := &x509.Certificate{
		RawSubject: []byte("leaf-subject"),
		RawIssuer:  []byte("issuer"),
		Raw:        rawCert,
		NotAfter:   notAfter,
		Subject:    pkix.Name{CommonName: "example.com"},
		Issuer:     pkix.Name{CommonName: "Test CA"},
		IsCA:       false,
	}
	state := &tls.ConnectionState{PeerCertificates: []*x509.Certificate{cert}}

	status, err := checkTls("test-service", state)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if status.ServiceName != "test-service" {
		t.Errorf("Expected ServiceName 'test-service', got %q", status.ServiceName)
	}

	wantSum := sha256.Sum256(rawCert)
	wantFingerprint := hex.EncodeToString(wantSum[:])
	if status.Fingerprint != wantFingerprint {
		t.Errorf("Expected fingerprint %q, got %q", wantFingerprint, status.Fingerprint)
	}
	if !status.Not_after.Equal(notAfter) {
		t.Errorf("Expected Not_after %v, got %v", notAfter, status.Not_after)
	}
	if status.Subject != "example.com" {
		t.Errorf("Expected subject 'example.com', got %q", status.Subject)
	}
	if status.Issuer != "Test CA" {
		t.Errorf("Expected issuer 'Test CA', got %q", status.Issuer)
	}
	if !status.Is_expired {
		t.Error("Expected Is_expired to be true for a certificate whose NotAfter is in the past")
	}
	if status.Chain == "" {
		t.Error("Expected Chain JSON to be populated")
	}
	if status.First_seen.IsZero() || status.Last_checked.IsZero() {
		t.Error("Expected First_seen and Last_checked to be set")
	}
}
