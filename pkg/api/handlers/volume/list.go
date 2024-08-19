package volume

import (
	"encoding/json"
	"net/http"

	"github.com/runfinch/finch-daemon/pkg/api/response"
)

func (h *handler) list(w http.ResponseWriter, r *http.Request) {
	// filters are JSON encoded value of the filters as a map[string][]string
	// https://docs.docker.com/engine/api/v1.40/#tag/Volume/operation/VolumeList
	rawJSONFilters := r.URL.Query().Get("filters")

	filtersSlice, err := rawJSONFiltersToSlice(rawJSONFilters)
	if err != nil && rawJSONFilters != "" {
		h.logger.Errorf("could not convert filters to JSON: %v", err)
		response.JSON(w, http.StatusInternalServerError, response.NewError(err))
	}

	resp, err := h.service.List(r.Context(), filtersSlice)
	if err != nil {
		response.JSON(w, http.StatusInternalServerError, response.NewError(err))
		return
	}
	response.JSON(w, http.StatusOK, resp)
}

// filterJSONRequest represents the raw JSON string passed as URL param to
// the volumes list API.

// An example encoded request for {"name": "test"} looks like:
//
//	GET /v1.41/volumes?filters=%7B%22name%22%3A%5B%22test%22%5D%7D'
type filterJSONRequest struct {
	Name     []string `json:"name"`
	Driver   []string `json:"driver"`
	Labels   []string `json:"labels"`
	Dangling []string `json:"dangling"`
}

// rawJSONFiltersToSlice converts a raw JSON URL object to a slice
// of individual filter expressions.
// e.g.:
//
//	{"name":["test", "bar"], "driver":["foo"]}
//
// becomes the string slice
//
//	{"name=test", "name=bar", "driver=foo"}
func rawJSONFiltersToSlice(rawJSONFilters string) ([]string, error) {
	filtersJSON := filterJSONRequest{}
	err := json.Unmarshal([]byte(rawJSONFilters), &filtersJSON)
	if err != nil {
		return nil, err
	}

	filters := []string{}
	danglingExprs := createFilterExpressions("dangling", filtersJSON.Dangling)
	driverExprs := createFilterExpressions("driver", filtersJSON.Driver)
	labelExprs := createFilterExpressions("label", filtersJSON.Labels)
	nameExprs := createFilterExpressions("name", filtersJSON.Name)

	filters = append(filters, danglingExprs...)
	filters = append(filters, driverExprs...)
	filters = append(filters, labelExprs...)
	filters = append(filters, nameExprs...)

	return filters, nil
}

func createFilterExpressions(key string, vals []string) []string {
	expressions := []string{}
	for _, val := range vals {
		expressions = append(expressions, key+"="+val)
	}
	return expressions
}
