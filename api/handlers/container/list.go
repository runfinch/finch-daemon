// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/containerd/containerd/v2/pkg/namespaces"
	ncTypes "github.com/containerd/nerdctl/v2/pkg/api/types"

	"github.com/runfinch/finch-daemon/api/response"
	"github.com/runfinch/finch-daemon/api/types"
)

const (
	allKey      = "all"
	limitKey    = "limit"
	sizeKey     = "size"
	filtersKey  = "filters"
	defaultAll  = false
	defaultSize = false
)

func (h *handler) list(w http.ResponseWriter, r *http.Request) {
	ctx := namespaces.WithNamespace(r.Context(), h.Config.Namespace)

	q := r.URL.Query()
	all, err := parseBoolQP(q, allKey, defaultAll)
	if err != nil {
		response.JSON(w, http.StatusBadRequest, response.NewErrorFromMsg(fmt.Sprintf("invalid query parameter \"all\": %s", err)))
		return
	}
	limit, err := parseIntQP(q, limitKey, 0)
	if err != nil {
		response.JSON(w, http.StatusBadRequest, response.NewErrorFromMsg(fmt.Sprintf("invalid query parameter \"limit\": %s", err)))
		return
	}
	// TODO: Size is not in the response so the parameter is not actually used. Add size to response later.
	size, err := parseBoolQP(q, sizeKey, defaultSize)
	if err != nil {
		response.JSON(w, http.StatusBadRequest, response.NewErrorFromMsg(fmt.Sprintf("invalid query parameter \"size\": %s", err)))
		return
	}
	filters, err := NerdctlFiltersFromAPIFilters(q)
	if err != nil {
		response.JSON(w, http.StatusBadRequest, response.NewErrorFromMsg(fmt.Sprintf("invalid query parameter \"filters\": %s", err)))
		return
	}

	globalOpt := ncTypes.GlobalCommandOptions(*h.Config)

	listOpts := ncTypes.ContainerListOptions{
		GOptions: globalOpt,
		All:      all,
		LastN:    limit,
		Truncate: true,
		Size:     size,
		Filters:  filters,
	}
	containers, err := h.service.List(ctx, listOpts)
	if err != nil {
		response.JSON(w, http.StatusInternalServerError, response.NewError(err))
		return
	}
	response.JSON(w, http.StatusOK, containers)
}

func parseBoolQP(q url.Values, key string, defaultV bool) (bool, error) {
	v := q.Get(key)
	if v == "" {
		return defaultV, nil
	} else {
		r, err := strconv.ParseBool(v)
		if err != nil {
			return false, err
		}
		return r, nil
	}
}

func parseIntQP(q url.Values, key string, defaultV int) (int, error) {
	v := q.Get(key)
	if v == "" {
		return defaultV, nil
	} else {
		r, err := strconv.ParseInt(v, 10, 0)
		if err != nil {
			return 0, err
		}
		return int(r), nil
	}
}

func NerdctlFiltersFromAPIFilters(query url.Values) ([]string, error) {
	filters, err := types.ParseFilterArgs(query)
	if err != nil {
		return nil, err
	}

	var ncFilters []string
	for filterType, filterList := range filters.ToLegacyFormat() {
		for _, f := range filterList {
			ncFilters = append(ncFilters, fmt.Sprintf("%s=%s", filterType, f))
		}
	}
	return ncFilters, nil
}
