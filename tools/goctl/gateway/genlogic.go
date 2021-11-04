package gateway

import (
	"bytes"
	"text/template"

	"github.com/micro-easy/go-zero/tools/goctl/api/util"
	ctlutil "github.com/micro-easy/go-zero/tools/goctl/util"
)

const logicTemplate = `package logic

import (
	{{.imports}}
)

type {{.logic}} struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func New{{.logic}}(ctx context.Context, svcCtx *svc.ServiceContext) {{.logic}} {
	return {{.logic}}{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *{{.logic}}) {{.function}}({{.request}}) {{.responseType}} {
	// todo: add your logic here and delete this line

	{{.returnString}}
}
`

func (g *GatewayGenerator) genLogic(dir string) error {
	// TODO methodName
	methodName := "methodName"
	fp, created, err := util.MaybeCreateFile(dir, handlerDir, methodName+"_handler.go")
	if err != nil {
		return err
	}
	if !created {
		return nil
	}
	defer fp.Close()

	// parentPkg, err := getParentPackage(dir)
	// if err != nil {
	// 	return err
	// }

	text, err := ctlutil.LoadTemplate(category, handlerTemplateFile, handlerTemplate)
	if err != nil {
		return err
	}

	buffer := new(bytes.Buffer)
	err = template.Must(template.New("handlerTemplate").Parse(text)).Execute(buffer,
		map[string]string{})
	if err != nil {
		return err
	}
	return nil
}
