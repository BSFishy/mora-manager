package wingman

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type GetConfigPointsResponse struct {
	ConfigPoints []ConfigPoint
}

func (a *app) handleConfigPoints(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	configPoints, err := a.wingman.GetConfigPoints(ctx)
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
