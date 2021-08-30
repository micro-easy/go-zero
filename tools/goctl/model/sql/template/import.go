package template

var (
	Imports = `import (
	"database/sql"
	"fmt"
	"strings"
	{{if .time}}"time"{{end}}

	"github.com/micro-easy/go-zero/core/stores/cache"
	"github.com/micro-easy/go-zero/core/stores/sqlc"
	"github.com/micro-easy/go-zero/core/stores/sqlx"
	"github.com/micro-easy/go-zero/core/stringx"
	"github.com/micro-easy/go-zero/tools/goctl/model/sql/builderx"
)
`
	ImportsNoCache = `import (
	"database/sql"
	"fmt"
	"strings"
	{{if .time}}"time"{{end}}

	"github.com/micro-easy/go-zero/core/stores/sqlc"
	"github.com/micro-easy/go-zero/core/stores/sqlx"
	"github.com/micro-easy/go-zero/core/stringx"
	"github.com/micro-easy/go-zero/tools/goctl/model/sql/builderx"
)
`
)
