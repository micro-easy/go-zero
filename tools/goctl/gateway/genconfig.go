package gateway

import (
	"bytes"
	"text/template"

	apiutil "github.com/micro-easy/go-zero/tools/goctl/api/util"
	ctlutil "github.com/micro-easy/go-zero/tools/goctl/util"
)

const (
	configFile     = "config.go"
	configTemplate = `package config

type Config struct {
	rest.RestConf
}
`
)

func (g *GatewayGenerator) genConfig(dir string) error {
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
	err = t.Execute(buffer, map[string]string{})
	if err != nil {
		return err
	}

	formatCode := formatCode(buffer.String())
	_, err = fp.WriteString(formatCode)
	return err
}
