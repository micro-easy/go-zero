package config

import (
	"github.com/micro-easy/go-zero/rest"
	"github.com/micro-easy/go-zero/zrpc"
)

type Config struct {
	rest.RestConf
	Rpc zrpc.RpcClientConf
}
