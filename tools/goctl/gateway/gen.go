package gateway

import (
	"errors"

	"github.com/urfave/cli"
)

// Gateway is to generate api service code from a proto file by specifying a proto file
func Gateway(c *cli.Context) error {
	src := c.String("proto")
	out := c.String("dir")
	pbPath := c.String("pbpath")
	pbClientPath := c.String("pbclientpath")
	if len(src) == 0 {
		return errors.New("missing -proto")
	}
	if len(out) == 0 {
		return errors.New("missing -dir")
	}
	if len(pbPath) == 0 {
		return errors.New("missing -pbpath")
	}

	if len(pbClientPath) == 0 {
		return errors.New("missing -pbclientpath")
	}

	return NewGatewayGenerator().Generate(src, out, pbPath, pbClientPath)
}
