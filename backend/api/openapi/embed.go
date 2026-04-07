// Package openapi embeds the spec file so it can be served by /openapi.yaml
// without depending on the working directory.
package openapi

import _ "embed"

//go:embed openapi.yaml
var SpecBytes []byte
