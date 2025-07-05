package wingman

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/BSFishy/mora-manager/point"
	"github.com/BSFishy/mora-manager/state"
)

type GetConfigPointsRequest struct {
	ModuleName string
	State      state.State
}

type GetConfigPointsResponse struct {
	ConfigPoints []point.Point
}

func (a *app) handleConfigPoints(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	var body GetConfigPointsRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return fmt.Errorf("decoding body: %w", err)
	}

	wingmanCtx := wingmanContext{
		client:     a.client,
		moduleName: body.ModuleName,
		state:      &body.State,
		registry:   a.registry,
	}

	configPoints, err := a.wingman.GetConfigPoints(ctx, &wingmanCtx)
	if err != nil {
		return fmt.Errorf("getting config points: %w", err)
	}

	err = json.NewEncoder(w).Encode(GetConfigPointsResponse{
		ConfigPoints: configPoints,
	})
	if err != nil {
		return fmt.Errorf("encoding config points: %w", err)
	}

	return nil
}
