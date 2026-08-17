package migrations

import "embed"

// FS contains the ordered SQL migration files for cmd/migrate.
//
//go:embed *.sql
var FS embed.FS
