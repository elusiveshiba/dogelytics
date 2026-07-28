// Package assets contains the web assets embedded in the Dogelytics binary.
package assets

import "embed"

// Files contains the dashboard images and favicon set.
//
//go:embed *.png *.svg favicons/*
var Files embed.FS
