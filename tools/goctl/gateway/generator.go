package gateway

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/micro-easy/go-zero/tools/goctl/util"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"

	goformat "go/format"

	"github.com/micro-easy/go-zero/tools/goctl/util/console"
	options "google.golang.org/genproto/googleapis/api/annotations"
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

	err = g.genHandler(abs)
	if err != nil {
		return err
	}

	err = g.genRoute(abs)
	if err != nil {
		return err
	}

	err = g.genLogic(abs)
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

func extractAPIOptions(meth *descriptorpb.MethodDescriptorProto) (*options.HttpRule, error) {
	if meth.Options == nil {
		return nil, nil
	}
	if !proto.HasExtension(meth.Options, options.E_Http) {
		return nil, nil
	}
	ext := proto.GetExtension(meth.Options, options.E_Http)
	opts, ok := ext.(*options.HttpRule)
	if !ok {
		return nil, fmt.Errorf("extension is %T; want an HttpRule", ext)
	}
	return opts, nil
}
