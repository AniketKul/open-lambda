package pmanager

import (
	"github.com/open-lambda/open-lambda/worker/config"
	sb "github.com/open-lambda/open-lambda/worker/sandbox"
)

type PoolManager interface {
	ForkEnter(sandbox sb.ContainerSandbox, req_pkgs []string) error
}

func InitPoolManager(config *config.Config) (pm PoolManager, err error) {
	if config.Pool == "basic" {
		if pm, err = NewBasicManager(config); err != nil {
			return nil, err
		}
	} else {
		pm = nil
	}

	return pm, nil
}
