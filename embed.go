package main

import "embed"

//go:embed docs/R1S/*.pdf docs/R1T/*.pdf
var docsFS embed.FS
