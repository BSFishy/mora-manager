package kube

import (
	"github.com/BSFishy/mora-manager/core"
	"github.com/BSFishy/mora-manager/model"
)

type KubeContext interface {
	core.HasClientSet
	core.HasModuleName
	model.HasUser
	model.HasEnvironment
}
