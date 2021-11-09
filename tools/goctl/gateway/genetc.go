package gateway

import (
	"bytes"
	"fmt"
	"text/template"

	apiutil "github.com/micro-easy/go-zero/tools/goctl/api/util"
	ctlutil "github.com/micro-easy/go-zero/tools/goctl/util"
)

const (
	etcDir      = "etc"
	etcTemplate = `Name: {{.ServiceName}}api
Host: 0.0.0.0
Port: 8000
{{.ServiceName}}:
  Etcd:
    Hosts:
      - 127.0.0.1:2379
    Key: {{.ServiceName}}rpc
`
)

func (g *GatewayGenerator) genEtc(dir, serviceName string) error {
	fp, created, err := apiutil.MaybeCreateFile(dir, etcDir, fmt.Sprintf("%sapi.yaml", serviceName))
	if err != nil {
		return err
	}
	if !created {
		return nil
	}
	defer fp.Close()

	text, err := ctlutil.LoadTemplate("api", etcTemplateFile, etcTemplate)
	if err != nil {
		return err
	}

	t := template.Must(template.New("etcTemplate").Parse(text))
	buffer := new(bytes.Buffer)
	err = t.Execute(buffer, map[string]string{
		"ServiceName": serviceName,
	})
	if err != nil {
		return err
	}

	formatCode := formatCode(buffer.String())
	_, err = fp.WriteString(formatCode)
	return err
}
