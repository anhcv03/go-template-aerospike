package gcommon

import (
	queryoptions "go.jtlabs.io/query"

	pb "gitlab.vht.vn/tt-kttt/lae-project/utm/utm-track-manager/pkg/pb"
)

/* Parse GRPC query options */
func ParseQueryOptions(so *pb.SearchOptions) queryoptions.Options {
	filter := map[string][]string{}
	page := map[string]int{}

	for k, v := range so.Filter {
		filter[k] = v.Str
	}

	for k, v := range so.Page {
		page[k] = int(v)
	}

	return queryoptions.Options{
		Fields: so.Fields,
		Filter: filter,
		Page:   page,
		Sort:   so.Sorts,
	}
}
