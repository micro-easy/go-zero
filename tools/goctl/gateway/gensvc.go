package gateway

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/micro-easy/go-zero/tools/goctl/api/util"
	ctlutil "github.com/micro-easy/go-zero/tools/goctl/util"
)

const (
	contextFilename = "servicecontext.go"
	contextTemplate = `package svc

import (
	{{.ImportPackages}}
	"github.com/micro-easy/go-zero/zrpc"
)

type ServiceContext struct {
	Config config.Config
	{{.ServiceName}} {{.PbClientPkg}}.{{.ServiceName}}
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config: c, 
		{{.ServiceName}}: {{.PbClientPkg}}.New{{.ServiceName}}(zrpc.MustNewClient(c.{{.ServiceName}})),
	}
}

`
)

func (g *GatewayGenerator) genSvc(dir, serviceName, pbClientPath string) error {
	fp, created, err := util.MaybeCreateFile(dir, contextDir, contextFilename)
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

	text, err := ctlutil.LoadTemplate(category, contextTemplateFile, contextTemplate)
	if err != nil {
		return err
	}

	t := template.Must(template.New("contextTemplate").Parse(text))
	buffer := new(bytes.Buffer)
	err = t.Execute(buffer, map[string]string{
		"ImportPackages": genSvcImports(parentPkg, pbClientPath),
		"ServiceName":    strings.Title(serviceName),
		"PbClientPkg":    pbClientPath[strings.LastIndex(pbClientPath, "/")+1:],
	})
	if err != nil {
		return err
	}

	formatCode := formatCode(buffer.String())
	_, err = fp.WriteString(formatCode)
	return err
}

func genSvcImports(parentPkg, pbClientPath string) string {
	var imports []string
	imports = append(imports, fmt.Sprintf("\"%s\"", ctlutil.JoinPackages(parentPkg, configDir)))
	imports = append(imports, fmt.Sprintf("\"%s\"", pbClientPath))
	return strings.Join(imports, "\n\t")
}
