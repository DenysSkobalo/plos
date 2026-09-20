package web

import "embed"

// DistFS містить скомпільовані стативні файли фронтенду з web/dist.
//
//go:embed all:dist
var DistFS embed.FS
