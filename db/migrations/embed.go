// Package migrations holds the goose migration set for the service_data
// database and embeds the .sql files so they ship inside the binary.
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
