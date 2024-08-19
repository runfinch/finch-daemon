package network

import (
	"net/http"

	"github.com/containerd/containerd/namespaces"
	"github.com/gorilla/mux"
	"github.com/runfinch/finch-daemon/pkg/api/response"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
	"github.com/sirupsen/logrus"
)

func (h *handler) remove(w http.ResponseWriter, r *http.Request) {
	ctx := namespaces.WithNamespace(r.Context(), h.config.Namespace)
	err := h.service.Remove(ctx, mux.Vars(r)["id"])
	if err != nil {
		var code int
		switch {
		case errdefs.IsNotFound(err):
			code = http.StatusNotFound
		case errdefs.IsForbiddenError(err):
			code = http.StatusForbidden
		default:
			code = http.StatusInternalServerError
		}
		logrus.Errorf("Network Remove API failed. Status code %d, Message: %s", code, err)
		response.SendErrorResponse(w, code, err)
		return
	}
	response.Status(w, http.StatusNoContent)
}
