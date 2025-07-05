package wingman

import (
	"context"
	"fmt"
	"net/http"

	"github.com/BSFishy/mora-manager/core"
	"github.com/BSFishy/mora-manager/expr"
	"github.com/BSFishy/mora-manager/function"
	"github.com/BSFishy/mora-manager/point"
	"github.com/BSFishy/mora-manager/state"
	"github.com/BSFishy/mora-manager/value"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

type Manager struct{}

func (m *Manager) GetWingmanManager() function.WingmanManager {
	return m
}

func (m *Manager) FindWingman(ctx context.Context, deps interface {
	core.HasUser
	core.HasEnvironment
	core.HasModuleName
	core.HasServiceName
	core.HasClientSet
},
) (*WingmanClient, error) {
	clientset := deps.GetClientset()
	user := deps.GetUser()
	environment := deps.GetEnvironment()
	moduleName := deps.GetModuleName()
	serviceName := deps.GetServiceName()

	namespace := fmt.Sprintf("%s-%s", user, environment)
	selector := labels.SelectorFromSet(map[string]string{
		"mora.enabled":     "true",
		"mora.user":        user,
		"mora.environment": environment,
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

		return &WingmanClient{
			client: http.Client{},
			url:    url,
		}, nil
	}

	if k8serror.IsNotFound(err) {
		return nil, nil
	}

	return nil, err
}

type functionContext struct {
	user        string
	environment string
	state       *state.State
	moduleName  string
}

func (f *functionContext) GetUser() string {
	return f.user
}

func (f *functionContext) GetEnvironment() string {
	return f.environment
}

func (f *functionContext) GetState() *state.State {
	return f.state
}

func (f *functionContext) GetModuleName() string {
	return f.moduleName
}

func (m *Manager) EvaluateFunction(ctx context.Context, deps interface {
	core.HasUser
	core.HasEnvironment
	core.HasClientSet
	state.HasState
}, name string, args expr.Args,
) (value.Value, []point.Point, error) {
	clientset := deps.GetClientset()
	user := deps.GetUser()
	environment := deps.GetEnvironment()

	namespace := fmt.Sprintf("%s-%s", user, environment)
	selector := labels.SelectorFromSet(map[string]string{
		"mora.enabled":     "true",
		"mora.user":        user,
		"mora.environment": environment,
		"mora.wingman":     "true",
	})

	services, err := clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: selector.String(),
	})

	if err == nil {
		items := services.Items
		if len(items) == 0 {
			return nil, nil, nil
		}

		for _, svc := range items {
			url := fmt.Sprintf("http://%s.%s:8080", svc.Name, svc.Namespace)
			client := &WingmanClient{
				client: http.Client{},
				url:    url,
			}

			svcDeps := &functionContext{
				user:        deps.GetUser(),
				environment: deps.GetEnvironment(),
				state:       deps.GetState(),
				moduleName:  svc.Labels["mora.module"],
			}

			val, points, err := client.GetFunction(ctx, svcDeps, name, args)
			if err != nil {
				return nil, nil, fmt.Errorf("evaluating wingman %s.%s: %w", svc.Name, svc.Namespace, err)
			}

			if val != nil || len(points) > 0 {
				return val, points, nil
			}
		}
	}

	if k8serror.IsNotFound(err) {
		return nil, nil, nil
	}

	return nil, nil, err
}

type HasManager interface {
	GetWingmanManager() *Manager
}
