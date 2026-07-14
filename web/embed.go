// Package web embeds the built frontend single-page application.
package web

import "embed"

//go:embed all:dist
var Dist embed.FS
