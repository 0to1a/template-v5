// Package db embeds the schema migrations into the server binary, so the
// server can apply them itself at startup and a deployed binary can never
// run against a schema it doesn't carry.
package db

import "embed"

//go:embed migrations/*.sql
var Migrations embed.FS
