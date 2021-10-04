package deployer

import (
	"context"

	"github.com/bhoriuchi/opa-bundle-server/core/logger"
)

var (
	Providers = map[string]NewDeployerFunc{}
)

type NewDeployerFunc func(opts *Options) (Deployer, error)

type Options struct {
	Name   string
	Config interface{}
	Logger logger.Logger
}

type Deployer interface {
	Deploy(ctx context.Context) (err error)
}
