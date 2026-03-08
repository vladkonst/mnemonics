// Package migrations provides the embedded SQL migration files.
package migrations

import "embed"

// FS contains all goose SQL migration files.
//
//go:embed *.sql
var FS embed.FS
