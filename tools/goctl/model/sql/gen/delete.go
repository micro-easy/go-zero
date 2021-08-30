package gen

import (
	"strings"

	"github.com/micro-easy/go-zero/core/collection"
	"github.com/micro-easy/go-zero/tools/goctl/model/sql/template"
	"github.com/micro-easy/go-zero/tools/goctl/util"
	"github.com/micro-easy/go-zero/tools/goctl/util/stringx"
)

func genDelete(table Table, withCache bool) (string, error) {
	keySet := collection.NewSet()
	keyVariableSet := collection.NewSet()
	for fieldName, key := range table.CacheKey {
		if fieldName == table.PrimaryKey.Name.Source() {
			keySet.AddStr(key.KeyExpression)
		} else {
			keySet.AddStr(key.DataKeyExpression)
		}
		keyVariableSet.AddStr(key.Variable)
	}

	camel := table.Name.ToCamel()
	text, err := util.LoadTemplate(category, deleteTemplateFile, template.Delete)
	if err != nil {
		return "", err
	}

	output, err := util.With("delete").
		Parse(text).
		Execute(map[string]interface{}{
			"upperStartCamelObject":     camel,
			"withCache":                 withCache,
			"containsIndexCache":        table.ContainsUniqueKey,
			"lowerStartCamelPrimaryKey": stringx.From(table.PrimaryKey.Name.ToCamel()).UnTitle(),
			"dataType":                  table.PrimaryKey.DataType,
			"keys":                      strings.Join(keySet.KeysStr(), "\n"),
			"originalPrimaryKey":        table.PrimaryKey.Name.Source(),
			"keyValues":                 strings.Join(keyVariableSet.KeysStr(), ", "),
		})
	if err != nil {
		return "", err
	}
	return output.String(), nil
}
