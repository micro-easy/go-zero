package gateway

import (
	"bytes"
	"text/template"

	"github.com/micro-easy/go-zero/tools/goctl/api/util"
	ctlutil "github.com/micro-easy/go-zero/tools/goctl/util"
)

const handlerTemplate = `package handler

import (
	"net/http"

	{{.ImportPackages}}
)

func {{.HandlerName}}(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		{{if .HasRequest}}var req types.{{.RequestType}}
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}{{end}}

		l := logic.New{{.LogicType}}(r.Context(), ctx)
		{{if .HasResp}}resp, {{end}}err := l.{{.Call}}({{if .HasRequest}}req{{end}})
		if err != nil {
			httpx.Error(w, err)
		} else {
			{{if .HasResp}}httpx.OkJson(w, resp){{else}}httpx.Ok(w){{end}}
		}
	}
}
`

func (g *GatewayGenerator) genHandler(dir string) error {
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
