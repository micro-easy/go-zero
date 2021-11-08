package gateway

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/micro-easy/go-zero/tools/goctl/api/util"
	"github.com/micro-easy/go-zero/tools/goctl/gateway/descriptor"
	ctlutil "github.com/micro-easy/go-zero/tools/goctl/util"
	"github.com/micro-easy/go-zero/tools/goctl/vars"
)

const logicTemplate = `package logic

import (
	{{.ImportPackages}}
)

type {{.b.Method.GetName}}Logic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func New{{.b.Method.GetName}}Logic (ctx context.Context, svcCtx *svc.ServiceContext) {{.b.Method.GetName}}Logic {
	return {{.b.Method.GetName}}{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *{{.b.Method.GetName}}Logic) {{.b.Method.GetName}}(req {{.b.Method.GetInputType.GetFullyQualifiedName}}) ({{.b.Method.GetOutputType.GetFullyQualifiedName}},error) {
	// todo: add your logic here and delete this line
	resp,err:=l.svcCtx.{{.b.Method.GetService.GetName}}.{{.b.Method.GetName}}(l.ctx,req)
	if err!=nil{
		return nil,err
	}
	return resp,nil
}
`

func (g *GatewayGenerator) genLogic(dir, pbImportPath string, meth *descriptor.MethodWithBindings) error {
	for _, binding := range meth.Bindings {
		methodName := meth.GetName()
		fp, created, err := util.MaybeCreateFile(dir, logicDir, methodName+"_logic.go")
		if err != nil {
			return err
		}
		if !created {
			return nil
		}
		defer fp.Close()

		parentPkg, err := getParentPackage(dir)
		if err != nil {
			return err
		}

		text, err := ctlutil.LoadTemplate(category, logicTemplateFile, logicTemplate)
		if err != nil {
			return err
		}

		buffer := new(bytes.Buffer)
		err = template.Must(template.New("logicTemplate").Parse(text)).Execute(buffer,
			map[string]interface{}{
				"ImportPackages": genLogicImports(parentPkg, pbImportPath),
				"b":              binding,
			})
		if err != nil {
			return err
		}

		formatCode := formatCode(buffer.String())
		_, err = fp.WriteString(formatCode)
		if err != nil {
			return err
		}
	}

	return nil
}

func genLogicImports(parentPkg, pbImportPath string) string {
	var imports []string
	imports = append(imports, `"context"`+"\n")
	imports = append(imports, fmt.Sprintf("\"%s\"", ctlutil.JoinPackages(parentPkg, contextDir)))
	imports = append(imports, fmt.Sprintf("\"%s\"\n", pbImportPath))

	imports = append(imports, fmt.Sprintf("\"%s/core/logx\"", vars.ProjectOpenSourceUrl))
	return strings.Join(imports, "\n\t")
}
