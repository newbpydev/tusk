// Package db owns the immutable forward migration inventory.
package db

import (
	"embed"
	"io/fs"
)

//go:embed migrations/*.sql schema_catalog.json
var migrations embed.FS

// Migrations exposes read-only migration assets, including their migrations/ prefix.
func Migrations() fs.FS { return migrations }
