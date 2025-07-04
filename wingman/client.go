package wingman

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/BSFishy/mora-manager/expr"
	"github.com/BSFishy/mora-manager/point"
	"github.com/BSFishy/mora-manager/util"
)

type client struct {
	Client http.Client
	Url    string
}

func (c *client) request(ctx context.Context, method, url string, body []byte) (*http.Response, error) {
	fullUrl := fmt.Sprintf("%s%s", c.Url, url)

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
		resp, err = c.Client.Do(req)
		if err == nil && resp.StatusCode == 200 {
			return resp, nil
		}

		statusCode := -1
		if resp != nil {
			statusCode = resp.StatusCode
		}

		util.LogFromCtx(ctx).Debug("retrying wingman request", requestGroup, slog.Group("response", "status", statusCode))
		time.Sleep(time.Second)
	}

	return nil, err
}

func (c *client) GetConfigPoints(ctx context.Context, deps WingmanContext) ([]point.Point, error) {
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

func (c *client) GetFunctions(ctx context.Context, deps WingmanContext) (map[string]expr.ExpressionFunction, error) {
	return nil, nil
}
