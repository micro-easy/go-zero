package svc

import "github.com/micro-easy/go-zero/zrpc"

type ServiceContext struct {
	Client zrpc.Client
}
