package gateway

import (
	"bytes"
	"strings"
	"text/template"

	apiutil "github.com/micro-easy/go-zero/tools/goctl/api/util"
	ctlutil "github.com/micro-easy/go-zero/tools/goctl/util"
)

const (
	configFile     = "config.go"
	configTemplate = `package config

	import (
		"github.com/micro-easy/go-zero/rest"
		"github.com/micro-easy/go-zero/zrpc"
	)

type Config struct {
	rest.RestConf
	{{.ServiceName}}  zrpc.RpcClientConf
}
`
)

func (g *GatewayGenerator) genConfig(dir, serviceName string) error {
	fp, created, err := apiutil.MaybeCreateFile(dir, configDir, configFile)
	if err != nil {
		return err
	}
	if !created {
		return nil
	}
	defer fp.Close()

	text, err := ctlutil.LoadTemplate(category, configTemplateFile, configTemplate)
	if err != nil {
		return err
	}

	t := template.Must(template.New("configTemplate").Parse(text))
	buffer := new(bytes.Buffer)
	err = t.Execute(buffer, map[string]string{
		"ServiceName": strings.Title(serviceName),
	})
	if err != nil {
		return err
	}

	formatCode := formatCode(buffer.String())
	_, err = fp.WriteString(formatCode)
	return err
}
