package wingman

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/BSFishy/mora-manager/expr"
	"github.com/BSFishy/mora-manager/point"
	"github.com/BSFishy/mora-manager/state"
)

type GetFunctionRequest struct {
	ModuleName  string
	State       state.State
	Username    string
	Environment string

	FunctionName string
	Args         expr.Args
}

type GetFunctionResponse struct {
	Found        bool
	ConfigPoints []point.Point
	Value        json.RawMessage
	State        state.State
}

func (a *app) handleFunction(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	var body GetFunctionRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return fmt.Errorf("decoding body: %w", err)
	}

	st := &body.State
	wingmanCtx := wingmanContext{
		client:      a.client,
		moduleName:  body.ModuleName,
		user:        body.Username,
		environment: body.Environment,
		state:       st,
		registry:    a.registry,
	}

	functions := a.wingman.GetFunctions()
	response := GetFunctionResponse{}
	for name, function := range functions {
		if name != body.FunctionName {
			continue
		}

		if function.IsInvalid(body.Args) {
			// TODO: handle errors properly
			return errors.New("invalid args")
		}

		val, cfp, err := function.Evaluate(ctx, &wingmanCtx, body.Args)
		if err != nil {
			return fmt.Errorf("evaluating function: %w", err)
		}

		response.Found = true
		response.State = *st
		response.ConfigPoints = cfp

		// TODO: properly handle nil val?
		json, err := json.Marshal(val)
		if err != nil {
			return fmt.Errorf("encoding return value as json: %w", err)
		}

		response.Value = json
		break
	}

	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		return fmt.Errorf("encoding config points: %w", err)
	}

	return nil
}
