package generator

import (
	"strings"

	"github.com/micro-easy/go-zero/tools/goctl/util/stringx"
)

func formatFilename(filename string) string {
	return strings.ToLower(stringx.From(filename).ToCamel())
}
