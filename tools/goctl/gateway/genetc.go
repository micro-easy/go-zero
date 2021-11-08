package gateway

import (
	"bytes"
	"fmt"
	"strconv"
	"text/template"

	apiutil "github.com/micro-easy/go-zero/tools/goctl/api/util"
	ctlutil "github.com/micro-easy/go-zero/tools/goctl/util"
)

const (
	etcDir      = "etc"
	etcTemplate = `Name: {{.serviceName}}
Host: {{.host}}
Port: {{.port}}
`
)

func (g *GatewayGenerator) genEtc(dir, serviceName string) error {
	fp, created, err := apiutil.MaybeCreateFile(dir, etcDir, fmt.Sprintf("%s.yaml", serviceName))
	if err != nil {
		return err
	}
	if !created {
		return nil
	}
	defer fp.Close()

	host := "0.0.0.0"
	port := strconv.Itoa(defaultPort)

	text, err := ctlutil.LoadTemplate("api", etcTemplateFile, etcTemplate)
	if err != nil {
		return err
	}

	t := template.Must(template.New("etcTemplate").Parse(text))
	buffer := new(bytes.Buffer)
	err = t.Execute(buffer, map[string]string{
		"serviceName": serviceName,
		"host":        host,
		"port":        port,
	})
	if err != nil {
		return err
	}

	formatCode := formatCode(buffer.String())
	_, err = fp.WriteString(formatCode)
	return err
}
