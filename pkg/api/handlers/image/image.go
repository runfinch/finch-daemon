package image

import (
	"context"
	"io"
	"net/http"

	"github.com/containerd/nerdctl/pkg/config"
	"github.com/containerd/nerdctl/pkg/inspecttypes/dockercompat"
	dockertypes "github.com/docker/cli/cli/config/types"
	"github.com/runfinch/finch-daemon/pkg/api/types"
	"github.com/runfinch/finch-daemon/pkg/flog"
)

//go:generate mockgen --destination=../../../mocks/mocks_image/imagesvc.go -package=mocks_image github.com/runfinch/finch-daemon/pkg/api/handlers/image Service
type Service interface {
	List(ctx context.Context) ([]types.ImageSummary, error)
	Pull(ctx context.Context, name, tag, platform string, authCfg *dockertypes.AuthConfig, outStream io.Writer) error
	Push(ctx context.Context, name, tag string, authCfg *dockertypes.AuthConfig, outStream io.Writer) (*types.PushResult, error)
	Remove(ctx context.Context, name string, force bool) (deleted, untagged []string, err error)
	Tag(ctx context.Context, srcImg string, repo, tag string) error
	Inspect(ctx context.Context, name string) (*dockercompat.Image, error)
}

func RegisterHandlers(r types.VersionedRouter, service Service, conf *config.Config, logger flog.Logger) {
	h := newHandler(service, conf, logger)

	r.SetPrefix("/images")
	r.HandleFunc("/create", h.pull, http.MethodPost)
	r.HandleFunc("/json", h.list, http.MethodGet)
	r.HandleFunc("/{name:.*}", h.remove, http.MethodDelete)
	r.HandleFunc("/{name:.*}/push", h.push, http.MethodPost)
	r.HandleFunc("/{name:.*}/tag", h.tag, http.MethodPost)
	r.HandleFunc("/{name:.*}/json", h.inspect, http.MethodGet)
}

func newHandler(service Service, conf *config.Config, logger flog.Logger) *handler {
	return &handler{
		service: service,
		Config:  conf,
		logger:  logger,
	}
}

type handler struct {
	service Service
	Config  *config.Config
	logger  flog.Logger
}
