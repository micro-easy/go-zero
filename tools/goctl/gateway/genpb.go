package gateway

import (
	"github.com/micro-easy/go-zero/tools/goctl/rpc/parser"
)

// dir/
// dir/etc
// dir/internal/config
// dir/internal/handler
// dir/internal/logic
// dir/internal/svc
// dir/internal/servicepb

// 只需要生成message的数据即可
func (g *GatewayGenerator) genPb(dir string, protoImportPath []string, proto parser.Proto) error {
	return nil
	// TODO service name
	// err := util.MkdirIfNotExist(dir +"/"+ internal+)
	// if err != nil {
	// 	return err
	// }
	// cw := new(bytes.Buffer)
	// base := filepath.Dir(proto.Src)
	// cw.WriteString("protoc ")
	// for _, ip := range protoImportPath {
	// 	cw.WriteString(" -I=" + ip)
	// }
	// cw.WriteString(" -I=" + base)
	// cw.WriteString(" " + proto.Name)
	// if strings.Contains(proto.GoPackage, "/") {
	// 	cw.WriteString(" --go_out=plugins=grpc:" + ctx.GetMain().Filename)
	// } else {
	// 	cw.WriteString(" --go_out=plugins=grpc:" + dir.Filename)
	// }
	// command := cw.String()
	// _, err := execx.Run(command, "")
	// return err
}
