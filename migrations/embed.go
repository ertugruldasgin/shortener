// Package migrations embed the SQL mihration files.
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
