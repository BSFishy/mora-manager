package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
)

// TODO: make this configurable, through env vars
const REPO_URL = "localhost:5000"

type ImagePushResponse struct {
	Image string `json:"image"`
}

func (a *App) imagePush(w http.ResponseWriter, r *http.Request) error {
	defer r.Body.Close()

	ctx := r.Context()
	user, _ := GetUser(ctx)

	query := r.URL.Query()
	environment := query.Get("environment")
	module := query.Get("module")
	image := query.Get("image")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("reading body: %w", err)
	}

	img, err := tarball.Image(func() (io.ReadCloser, error) { return io.NopCloser(bytes.NewReader(body)), nil }, nil)
	if err != nil {
		return fmt.Errorf("creating image: %w", err)
	}

	pushTag, err := name.NewTag(fmt.Sprintf("%s/%s_%s/%s_%s", REPO_URL, user.Username, environment, module, image))
	if err != nil {
		return fmt.Errorf("creating push tag: %w", err)
	}

	if err = remote.Push(pushTag, img); err != nil {
		return fmt.Errorf("pushing image: %w", err)
	}

	response := ImagePushResponse{
		Image: pushTag.String(),
	}

	data, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("encoding response: %w", err)
	}

	_, err = w.Write(data)
	if err != nil {
		return fmt.Errorf("writing response: %w", err)
	}

	return nil
}
