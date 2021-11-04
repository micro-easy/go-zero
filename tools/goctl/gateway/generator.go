package gateway

import (
	"os/exec"
	"path/filepath"

	"github.com/micro-easy/go-zero/tools/goctl/util"

	goformat "go/format"

	"github.com/micro-easy/go-zero/tools/goctl/util/console"
)

type GatewayGenerator struct {
}

func NewGatewayGenerator() *GatewayGenerator {
	return &GatewayGenerator{}
}

func (g *GatewayGenerator) Generate(src, target string, protoImportPath []string) error {
	abs, err := filepath.Abs(target)
	if err != nil {
		return err
	}

	err = util.MkdirIfNotExist(abs)
	if err != nil {
		return err
	}

	err = g.Prepare()
	if err != nil {
		return err
	}

	// projectCtx, err := ctx.Prepare(abs)
	// if err != nil {
	// 	return err
	// }

	// p := parser.NewDefaultProtoParser()
	// proto, err := p.Parse(src)
	// if err != nil {
	// 	return err
	// }

	// dirCtx, err := mkdir(projectCtx, proto)
	// if err != nil {
	// 	return err
	// }

	err = g.genEtc(abs)
	if err != nil {
		return err
	}

	// err = g.g.GenPb(dirCtx, protoImportPath, proto)
	// if err != nil {
	// 	return err
	// }

	err = g.genConfig(abs)
	if err != nil {
		return err
	}

	err = g.genSvc(abs)
	if err != nil {
		return err
	}

	// err = g.g.GenLogic(dirCtx, proto)
	// if err != nil {
	// 	return err
	// }

	// err = g.g.GenServer(dirCtx, proto)
	// if err != nil {
	// 	return err
	// }

	err = g.genMain(abs)
	if err != nil {
		return err
	}

	// err = g.g.GenCall(dirCtx, proto)

	console.NewColorConsole().MarkDone()

	return err
}

func (g *GatewayGenerator) Prepare() error {
	_, err := exec.LookPath("go")
	if err != nil {
		return err
	}

	_, err = exec.LookPath("protoc")
	if err != nil {
		return err
	}

	_, err = exec.LookPath("protoc-gen-go")
	return err
}

func formatCode(code string) string {
	ret, err := goformat.Source([]byte(code))
	if err != nil {
		return code
	}

	return string(ret)
}
