package wingman

import (
	"context"
	"fmt"
	"net/http"

	"github.com/BSFishy/mora-manager/core"
	"github.com/BSFishy/mora-manager/model"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

type Manager struct{}

func (m *Manager) FindWingman(ctx context.Context, deps interface {
	model.HasUser
	model.HasEnvironment
	core.HasModuleName
	core.HasServiceName
	core.HasClientSet
},
) (Wingman, error) {
	clientset := deps.GetClientset()
	user := deps.GetUser()
	environment := deps.GetEnvironment()
	moduleName := deps.GetModuleName()
	serviceName := deps.GetServiceName()

	namespace := fmt.Sprintf("%s-%s", user.Username, environment.Slug)
	selector := labels.SelectorFromSet(map[string]string{
		"mora.enabled":     "true",
		"mora.user":        user.Username,
		"mora.environment": environment.Slug,
		"mora.module":      moduleName,
		"mora.service":     serviceName,
		"mora.wingman":     "true",
	})

	services, err := clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: selector.String(),
	})

	if err == nil {
		items := services.Items
		if len(items) == 0 {
			return nil, nil
		}

		if len(items) != 1 {
			return nil, fmt.Errorf("invalid number of services matched wingman: %d", len(items))
		}

		svc := items[0]
		url := fmt.Sprintf("http://%s.%s:8080", svc.Name, svc.Namespace)

		return &client{
			Client: http.Client{},
			Url:    url,
		}, nil
	}

	if k8serror.IsNotFound(err) {
		return nil, nil
	}

	return nil, err
}

type HasManager interface {
	GetWingmanManager() *Manager
}
