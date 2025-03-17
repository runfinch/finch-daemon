// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// Filters represents a collection of filter types and their values.
type Filters struct {
	filterSet map[string]map[string]bool
}

// FilterKeyVal represents a single key-value pair for a filter.
type FilterKeyVal struct {
	Key   string
	Value string
}

// newFilters creates a new Filters instance with the provided filter key-value pairs.
func newFilters(filtersList ...FilterKeyVal) Filters {
	filters := Filters{filterSet: map[string]map[string]bool{}}
	for _, arg := range filtersList {
		filters.add(arg.Key, arg.Value)
	}
	return filters
}

// getFiltersKeys returns all values for a given filter key.
func (filters Filters) getFiltersKeys(key string) []string {
	values := filters.filterSet[key]
	if values == nil {
		return make([]string, 0)
	}
	slice := make([]string, 0, len(values))
	for key := range values {
		slice = append(slice, key)
	}
	return slice
}

// add inserts a new key-value pair into the filter set.
func (filters Filters) add(key, value string) {
	if _, ok := filters.filterSet[key]; ok {
		filters.filterSet[key][value] = true
	} else {
		filters.filterSet[key] = map[string]bool{value: true}
	}
}

// getFilterFromLegacy converts the legacy filter format (map[string][]string)
// to the new filter format (map[string]map[string]bool),
// legacy filter format is currently the default format in the API spec.
func getFilterFromLegacy(d map[string][]string) map[string]map[string]bool {
	m := map[string]map[string]bool{}
	for k, v := range d {
		values := map[string]bool{}
		for _, vv := range v {
			values[vv] = true
		}
		m[k] = values
	}
	return m
}

// getFilterJSON parses a JSON string into a Filters struct.
func getFilterJSON(p string) (Filters, error) {
	filters := newFilters()

	if p == "" {
		return filters, nil
	}

	raw := []byte(p)
	err := json.Unmarshal(raw, &filters)
	if err == nil {
		return filters, nil
	}

	// Fallback to parsing arguments in the legacy slice format
	deprecated := map[string][]string{}
	if legacyErr := json.Unmarshal(raw, &deprecated); legacyErr != nil {
		return filters, legacyErr
	}

	filters.filterSet = getFilterFromLegacy(deprecated)
	return filters, nil
}

// UnmarshalJSON implements the json.Unmarshaler interface for Filters.
func (filters Filters) UnmarshalJSON(raw []byte) error {
	return json.Unmarshal(raw, &filters.filterSet)
}

// ToLegacyFormat converts the Filters struct to the legacy format (map[string][]string).
func (filters Filters) ToLegacyFormat() map[string][]string {
	result := make(map[string][]string)
	for key := range filters.filterSet {
		values := filters.getFiltersKeys(key)
		if len(values) > 0 {
			result[key] = values
		}
	}
	return result
}

// ParseFilterArgs extracts and parses filters from URL query parameters.
func ParseFilterArgs(query url.Values) (Filters, error) {
	filterQuery := query.Get("filters")
	if filterQuery == "" {
		return newFilters(), nil
	}

	filters, err := getFilterJSON(filterQuery)
	if err != nil {
		return Filters{}, fmt.Errorf("error parsing filters: %v", err)
	}

	return filters, nil
}
