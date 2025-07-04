package core

import "k8s.io/client-go/kubernetes"

type HasClientSet interface {
	GetClientset() kubernetes.Interface
}
