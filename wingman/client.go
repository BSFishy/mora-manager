package wingman

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/BSFishy/mora-manager/core"
	"github.com/BSFishy/mora-manager/expr"
	"github.com/BSFishy/mora-manager/point"
	"github.com/BSFishy/mora-manager/util"
	"github.com/BSFishy/mora-manager/value"
)

type WingmanClient struct {
	client http.Client
	url    string
}

func (c *WingmanClient) request(ctx context.Context, method, url string, body []byte) (*http.Response, error) {
	fullUrl := fmt.Sprintf("%s%s", c.url, url)

	requestGroup := slog.Group("request", "method", method, "url", url)
	util.LogFromCtx(ctx).Debug("querying wingman", requestGroup)

	var err error
	for range 10 {
		var req *http.Request
		req, err = http.NewRequestWithContext(ctx, method, fullUrl, bytes.NewReader(body))
		if err != nil {
			return nil, fmt.Errorf("creating request: %w", err)
		}

		var resp *http.Response
		resp, err = c.client.Do(req)
		if err == nil && resp.StatusCode == 200 {
			return resp, nil
		}

		statusCode := -1
		if resp != nil {
			statusCode = resp.StatusCode
		}

		if err == nil {
			err = fmt.Errorf("invalid status code: %d", statusCode)
		}

		util.LogFromCtx(ctx).Debug("retrying wingman request", requestGroup, slog.Group("response", "status", statusCode))
		time.Sleep(time.Second)
	}

	return nil, err
}

func (c *WingmanClient) GetConfigPoints(ctx context.Context, deps WingmanContext) ([]point.Point, error) {
	state := deps.GetState()
	moduleName := deps.GetModuleName()

	bodyData := GetConfigPointsRequest{
		ModuleName: moduleName,
		State:      *state,
	}

	body, err := json.Marshal(bodyData)
	if err != nil {
		return nil, fmt.Errorf("encoding body: %w", err)
	}

	resp, err := c.request(ctx, http.MethodPost, "/api/v1/config-point", body)
	if err != nil {
		return nil, fmt.Errorf("getting endpoint: %w", err)
	}

	var data GetConfigPointsResponse
	if err = json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("decoding body: %w", err)
	}

	points := make([]point.Point, len(data.ConfigPoints))
	for i, point := range data.ConfigPoints {
		point.Fill(deps)

		points[i] = point
	}

	return points, nil
}

func (c *WingmanClient) GetFunction(ctx context.Context, deps interface {
	WingmanContext
	core.HasUser
	core.HasEnvironment
}, name string, args expr.Args,
) (value.Value, []point.Point, error) {
	state := deps.GetState()
	moduleName := deps.GetModuleName()

	bodyData := GetFunctionRequest{
		ModuleName:  moduleName,
		State:       *state,
		Username:    deps.GetUser(),
		Environment: deps.GetEnvironment(),

		FunctionName: name,
		Args:         args,
	}

	body, err := json.Marshal(bodyData)
	if err != nil {
		return nil, nil, fmt.Errorf("encoding body: %w", err)
	}

	resp, err := c.request(ctx, http.MethodPost, "/api/v1/function", body)
	if err != nil {
		return nil, nil, fmt.Errorf("getting endpoint: %w", err)
	}

	var data GetFunctionResponse
	if err = json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, nil, fmt.Errorf("decoding body: %w", err)
	}

	if !data.Found {
		return nil, nil, nil
	}

	if len(data.ConfigPoints) > 0 {
		return nil, data.ConfigPoints, nil
	}

	value, err := value.Unmarshal(data.Value)
	if err != nil {
		return nil, nil, fmt.Errorf("decoding return value: %w", err)
	}

	return value, nil, nil
}
