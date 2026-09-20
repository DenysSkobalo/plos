package migrations

import "embed"

// FS містить усі SQL міграції з каталогу migrations/.
//
//go:embed *.sql
var FS embed.FS
