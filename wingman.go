package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/BSFishy/mora-manager/config"
	"github.com/BSFishy/mora-manager/model"
	"github.com/BSFishy/mora-manager/util"
	"github.com/BSFishy/mora-manager/wingman"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

type RunwayWingman struct {
	Client http.Client
	Url    string
}

func (a *App) FindWingman(ctx context.Context) (wingman.Wingman, error) {
	user := util.Has(model.GetUser(ctx))
	environment := util.Has(model.GetEnvironment(ctx))
	moduleName := util.Has(util.GetModuleName(ctx))
	serviceName := util.Has(util.GetServiceName(ctx))

	namespace := fmt.Sprintf("%s-%s", user.Username, environment.Slug)
	selector := labels.SelectorFromSet(map[string]string{
		"mora.enabled":     "true",
		"mora.user":        user.Username,
		"mora.environment": environment.Slug,
		"mora.module":      moduleName,
		"mora.service":     serviceName,
		"mora.wingman":     "true",
	})

	services, err := a.clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{
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

		return &RunwayWingman{
			Client: http.Client{},
			Url:    url,
		}, nil
	}

	if k8serror.IsNotFound(err) {
		return nil, nil
	}

	return nil, err
}

func (r *RunwayWingman) request(method, url string, body []byte) (*http.Response, error) {
	fullUrl := fmt.Sprintf("%s%s", r.Url, url)

	var err error
	for range 10 {
		var req *http.Request
		req, err = http.NewRequest(method, fullUrl, bytes.NewReader(body))
		if err != nil {
			return nil, fmt.Errorf("creating request: %w", err)
		}

		var resp *http.Response
		resp, err = r.Client.Do(req)
		if err == nil && resp.StatusCode == 200 {
			return resp, nil
		}

		time.Sleep(time.Second)
	}

	return nil, err
}

func (r *RunwayWingman) GetConfigPoints(ctx context.Context) ([]config.Point, error) {
	state := util.Has(GetState(ctx))
	moduleName := util.Has(util.GetModuleName(ctx))

	bodyData := wingman.GetConfigPointsRequest{
		ModuleName: moduleName,
		State:      *state,
	}

	body, err := json.Marshal(bodyData)
	if err != nil {
		return nil, fmt.Errorf("encoding body: %w", err)
	}

	resp, err := r.request(http.MethodPost, "/api/v1/config-point", body)
	if err != nil {
		return nil, fmt.Errorf("getting endpoint: %w", err)
	}

	var data wingman.GetConfigPointsResponse
	if err = json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("decoding body: %w", err)
	}

	return data.ConfigPoints, nil
}
