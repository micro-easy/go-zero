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

const handlerTemplate = `package handler

import (
	"net/http"

	{{.ImportPackages}}
)

func {{.b.Method.GetName}}V{{.b.Index}}Handler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req {{.b.Method.GetInputType.GetFullyQualifiedName}}
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}

		{{if .b.PathParams}}
	var (
		val string
{{- if .b.HasEnumPathParam}}
		e int32
{{- end}}
{{- if .b.HasRepeatedEnumPathParam}}
		es []int32
{{- end}}
		ok bool
		err error
		_ = err
	)
	{{$binding := .b}}
	{{range $param := .b.PathParams}}
	{{$enum := $binding.LookupEnum $param}}
	val, ok = pathParams[{{$param | printf "%q"}}]
	if !ok {
		httpx.Error(w,fmt.Errorf("missing parameter %s", {{$param | printf "%q"}}))
		return  
	}
{{if $param.IsNestedProto3}}
	err = runtime.PopulateFieldFromPath(&req, {{$param | printf "%q"}}, val)
	if err != nil {
		httpx.Error(w,fmt.Errorf("type mismatch, parameter: %s, error: %v", {{$param | printf "%q"}}, err)
		return 
	}
	{{if $enum}}
		e{{if $param.IsRepeated}}s{{end}}, err = {{$param.ConvertFuncExpr}}(val{{if $param.IsRepeated}}, {{$binding.GetRepeatedPathParamSeparator | printf "%c" | printf "%q"}}{{end}}, {{$enum.GetFullyQualifiedName}}_value)
		if err != nil {
			httpx.Error(w,fmt.Errorf( "could not parse path as enum value, parameter: %s, error: %v", {{$param | printf "%q"}}, err)
			return 
		}
	{{end}}
{{else if $enum}}
	e{{if $param.IsRepeated}}s{{end}}, err = {{$param.ConvertFuncExpr}}(val{{if $param.IsRepeated}}, {{$binding.GetRepeatedPathParamSeparator | printf "%c" | printf "%q"}}{{end}}, {{$enum.GetFullyQualifiedName}}_value)
	if err != nil {
		httpx.Error("type mismatch, parameter: %s, error: %v", {{$param | printf "%q"}}, err)
		return 
	}
{{else}}
	{{$param.AssignableExpr "req"}}, err = {{$param.ConvertFuncExpr}}(val{{if $param.IsRepeated}}, {{$binding.GetRepeatedPathParamSeparator | printf "%c" | printf "%q"}}{{end}})
	if err != nil {
		httpx.Error("type mismatch, parameter: %s, error: %v", {{$param | printf "%q"}}, err)
		return 
	}
{{end}}
{{if and $enum $param.IsRepeated}}
	s := make([]{{$enum.GetFullyQualifiedName}}, len(es))
	for i, v := range es {
		s[i] = {{$enum.GetFullyQualifiedName}}(v)
	}
	{{$param.AssignableExpr "req"}} = s
{{else if $enum}}
	{{$param.AssignableExpr "req"}} = {{$enum.GetFullyQualifiedName}}(e)
{{end}}
	{{end}}
{{end}}

		l := logic.New{{.b.Method.GetName}}ApiLogic(r.Context(), ctx)
		resp, err := l.{{.b.Method.GetName}}Api(req)
		if err != nil {
			httpx.Error(w, err)
		} else {
			httpx.OkJson(w, resp)
		}
	}
}
`

func (g *GatewayGenerator) genHandler(dir, pbImportPath string, meth *descriptor.MethodWithBindings) error {
	for _, binding := range meth.Bindings {
		methodName := meth.GetName()
		fp, created, err := util.MaybeCreateFile(dir, handlerDir, methodName+"_handler.go")
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

		text, err := ctlutil.LoadTemplate(category, handlerTemplateFile+"notmatch", handlerTemplate)
		if err != nil {
			return err
		}

		buffer := new(bytes.Buffer)
		err = template.Must(template.New("handlerTemplate").Parse(text)).Execute(buffer,
			map[string]interface{}{
				"ImportPackages": genHandlerImports(parentPkg, pbImportPath),
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

func genHandlerImports(parentPkg, pbImportPath string) string {
	var imports []string
	imports = append(imports, fmt.Sprintf("\"%s\"",
		ctlutil.JoinPackages(parentPkg, logicDir)))
	imports = append(imports, fmt.Sprintf("\"%s\"", ctlutil.JoinPackages(parentPkg, contextDir)))
	// TODO 这里要填入pb的import路径
	// if len(route.RequestType.Name) > 0 {
	// 	imports = append(imports, fmt.Sprintf("\"%s\"\n", ctlutil.JoinPackages(parentPkg, typesDir)))
	// }
	imports = append(imports, fmt.Sprintf("\"%s\"\n", pbImportPath))
	imports = append(imports, fmt.Sprintf("\"%s/rest/httpx\"", vars.ProjectOpenSourceUrl))

	return strings.Join(imports, "\n\t")
}
