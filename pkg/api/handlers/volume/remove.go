package volume

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/runfinch/finch-daemon/pkg/api/response"
	"github.com/runfinch/finch-daemon/pkg/errdefs"
)

// remove handler deletes a volume if exists and not being used by any container
func (h *handler) remove(w http.ResponseWriter, r *http.Request) {
	volName := mux.Vars(r)["name"]
	force, err := strconv.ParseBool(r.URL.Query().Get("force"))
	if err != nil {
		force = false
	}

	err = h.service.Remove(r.Context(), volName, force)
	if err != nil {
		var code int
		switch {
		case errdefs.IsNotFound(err):
			code = http.StatusNotFound
		case errdefs.IsConflict(err):
			code = http.StatusConflict
		default:
			code = http.StatusInternalServerError
		}
		response.SendErrorResponse(w, code, err)
		return
	}
	response.Status(w, http.StatusNoContent)
}
