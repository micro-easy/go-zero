package gateway

import (
	"errors"

	"github.com/urfave/cli"
)

// Gateway is to generate api service code from a proto file by specifying a proto file
func Gateway(c *cli.Context) error {
	src := c.String("proto")
	out := c.String("dir")
	protoImportPath := c.StringSlice("proto_path")
	if len(src) == 0 {
		return errors.New("missing -proto")
	}
	if len(out) == 0 {
		return errors.New("missing -dir")
	}
	return NewGatewayGenerator().Generate(src, out, protoImportPath)
}
