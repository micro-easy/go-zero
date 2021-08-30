package template

var Error = `package {{.pkg}}

import "github.com/micro-easy/go-zero/core/stores/sqlx"

var ErrNotFound = sqlx.ErrNotFound
`
