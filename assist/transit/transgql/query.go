package transgql

import (
	"context"

	"go.stevenxie.me/api/v2/assist/transit"
	"go.stevenxie.me/api/v2/location/locgql"
)

// NewQuery creates a new Query.
func NewQuery(svc transit.Service) Query {
	return Query{svc: svc}
}

// A Query resolves queries for transit-related data.
type Query struct {
	svc transit.Service
}

// FindDepartures forwards a definition.
func (q Query) FindDepartures(
	ctx context.Context,
	route string,
	coords locgql.CoordinatesInput,
	radius *int,
	singleSet *bool,
) ([]transit.NearbyDeparture, error) {
	return q.svc.FindDepartures(
		ctx,
		route, locgql.CoordinatesFromInput(coords),
		func(opt *transit.FindDeparturesOptions) {
			opt.FuzzyMatch = true
			opt.GroupByStation = true
			if singleSet != nil {
				opt.SingleSet = *singleSet
			}
			if radius != nil {
				opt.Radius = *radius
			}
		},
	)
}

// NearbyTransports forwards a definition.
func (q Query) NearbyTransports(
	ctx context.Context,
	coords locgql.CoordinatesInput,
	radius *int,
	limit *int,
) ([]transit.Transport, error) {
	return q.svc.NearbyTransports(
		ctx,
		locgql.CoordinatesFromInput(coords),
		func(opt *transit.NearbyTransportsOptions) {
			if radius != nil {
				opt.Radius = *radius
			}
			if limit != nil {
				opt.Limit = *limit
			}
		},
	)
}
