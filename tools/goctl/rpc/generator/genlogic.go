package generator

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/micro-easy/go-zero/core/collection"
	"github.com/micro-easy/go-zero/tools/goctl/rpc/parser"
	"github.com/micro-easy/go-zero/tools/goctl/util"
	"github.com/micro-easy/go-zero/tools/goctl/util/stringx"
)

const (
	logicTemplate = `package logic

import (
	"context"

	{{.imports}}

	"github.com/micro-easy/go-zero/core/logx"
)

type {{.logicName}} struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func New{{.logicName}}(ctx context.Context,svcCtx *svc.ServiceContext) *{{.logicName}} {
	return &{{.logicName}}{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}
{{.functions}}
`
	logicFunctionTemplate = `{{if .hasComment}}{{.comment}}{{end}}
func (l *{{.logicName}}) {{.method}} (in {{.request}}) ({{.response}}, error) {
	// todo: add your logic here and delete this line
	
	return &{{.responseType}}{}, nil
}
`
)

func (g *defaultGenerator) GenLogic(ctx DirContext, proto parser.Proto) error {
	dir := ctx.GetLogic()
	for _, rpc := range proto.Service.RPC {
		filename := filepath.Join(dir.Filename, formatFilename(rpc.Name+"_logic")+".go")
		functions, err := g.genLogicFunction(proto.PbPackage, rpc)
		if err != nil {
			return err
		}

		imports := collection.NewSet()
		imports.AddStr(fmt.Sprintf(`"%v"`, ctx.GetSvc().Package))
		imports.AddStr(fmt.Sprintf(`"%v"`, ctx.GetPb().Package))
		text, err := util.LoadTemplate(category, logicTemplateFileFile, logicTemplate)
		if err != nil {
			return err
		}
		err = util.With("logic").GoFmt(true).Parse(text).SaveTo(map[string]interface{}{
			"logicName": fmt.Sprintf("%sLogic", stringx.From(rpc.Name).ToCamel()),
			"functions": functions,
			"imports":   strings.Join(imports.KeysStr(), util.NL),
		}, filename, false)
		if err != nil {
			return err
		}
	}
	return nil
}

func (g *defaultGenerator) genLogicFunction(goPackage string, rpc *parser.RPC) (string, error) {
	var functions = make([]string, 0)
	text, err := util.LoadTemplate(category, logicFuncTemplateFileFile, logicFunctionTemplate)
	if err != nil {
		return "", err
	}

	logicName := stringx.From(rpc.Name + "_logic").ToCamel()
	comment := parser.GetComment(rpc.Doc())
	buffer, err := util.With("fun").Parse(text).Execute(map[string]interface{}{
		"logicName":    logicName,
		"method":       parser.CamelCase(rpc.Name),
		"request":      fmt.Sprintf("*%s.%s", goPackage, parser.CamelCase(rpc.RequestType)),
		"response":     fmt.Sprintf("*%s.%s", goPackage, parser.CamelCase(rpc.ReturnsType)),
		"responseType": fmt.Sprintf("%s.%s", goPackage, parser.CamelCase(rpc.ReturnsType)),
		"hasComment":   len(comment) > 0,
		"comment":      comment,
	})
	if err != nil {
		return "", err
	}

	functions = append(functions, buffer.String())
	return strings.Join(functions, util.NL), nil
}
