package gateway

import (
	"bytes"
	"html/template"

	"github.com/micro-easy/go-zero/tools/goctl/api/util"
	ctlutil "github.com/micro-easy/go-zero/tools/goctl/util"
)

const (
	contextFilename = "servicecontext.go"
	contextTemplate = `package svc

import (
	{{.configImport}}
)

type ServiceContext struct {
	Config {{.config}}
	{{.middleware}}
}

func NewServiceContext(c {{.config}}) *ServiceContext {
	return &ServiceContext{
		Config: c, 
		{{.middlewareAssignment}}
	}
}

`
)

func (g *GatewayGenerator) genSvc(dir string) error {
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

	var configImport = "\"" + ctlutil.JoinPackages(parentPkg, configDir) + "\""

	t := template.Must(template.New("contextTemplate").Parse(text))
	buffer := new(bytes.Buffer)
	err = t.Execute(buffer, map[string]string{
		"configImport": configImport,
		"config":       "config.Config",
	})
	if err != nil {
		return err
	}

	formatCode := formatCode(buffer.String())
	_, err = fp.WriteString(formatCode)
	return err
}
