package wingman

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/BSFishy/mora-manager/state"
)

type GetConfigPointsRequest struct {
	State state.State
}

type GetConfigPointsResponse struct {
	ConfigPoints []ConfigPoint
}

func (a *app) handleConfigPoints(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	var body GetConfigPointsRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return fmt.Errorf("decoding body: %w", err)
	}

	configPoints, err := a.wingman.GetConfigPoints(ctx, body.State)
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
