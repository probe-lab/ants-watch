package db

import "embed"

// Migrations holds the ClickHouse schema migrations, embedded so they can be
// applied in-process via go-commons' db.ClickHouseMigrationsConfig.Apply.
//
//go:embed migrations/*.sql
var Migrations embed.FS
