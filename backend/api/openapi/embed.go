// Package openapi embeds the bundled spec so it can be served at /openapi.yaml
// without depending on the working directory.
//
// Source of truth: openapi.root.yaml + domains/* + components/*
// Bundle command: make api-bundle  (writes dist/openapi.yaml)
package openapi

import _ "embed"

//go:embed dist/openapi.yaml
var SpecBytes []byte
