package kube

import (
	"github.com/BSFishy/mora-manager/core"
)

type KubeContext interface {
	core.HasClientSet
	core.HasModuleName
	core.HasUser
	core.HasEnvironment
}
