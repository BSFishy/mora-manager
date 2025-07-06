package main

import (
	"context"
	"time"

	"github.com/BSFishy/mora-manager/model"
	"github.com/BSFishy/mora-manager/util"
)

// poll the db to see if the current deployment has been cancelled. if so,
// cancel the deployment
//
// TODO: there could still technically be some contention here, where a new
// deployment starts trying to deploy stuff while the previous one is still in
// the 2 second tick period and they try to fight over deploying the same
// resource. this is probably a pretty niche edge case but still a problem
func (a *App) handleDeployCancel(ctx context.Context, cancel context.CancelFunc, d *model.Deployment) {
	logger := util.LogFromCtx(ctx)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// context cancelled somewhere else
			return
		case <-ticker.C:
			isCancelled, err := d.IsCancelled(ctx, a.db)
			if err != nil {
				logger.Error("failed to check if deployment is cancelled", "err", err)
				continue
			}

			if isCancelled {
				logger.Info("ending cancelled deployment")
				cancel()
				return
			}
		}
	}
}
