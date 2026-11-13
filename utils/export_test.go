package utils

import "net/http"

// SetHTTPClientTransport swaps the transport used by notifier requests so
// external-package tests can redirect outbound traffic. Returns a restore func.
func SetHTTPClientTransport(rt http.RoundTripper) func() {
	old := httpClient.Transport
	httpClient.Transport = rt
	return func() { httpClient.Transport = old }
}
