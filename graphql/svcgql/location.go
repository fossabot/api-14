package svcgql

import (
	"go.stevenxie.me/api/graphql"
	"go.stevenxie.me/api/location/locgql"
)

type locationResolvers struct {
	address locgql.AddressResolver
	place   locgql.PlaceResolver
	seg     locgql.HistorySegmentResolver
}

func (res locationResolvers) Place() graphql.PlaceResolver     { return res.place }
func (res locationResolvers) Address() graphql.AddressResolver { return res.address }
func (res locationResolvers) LocationHistorySegment() graphql.LocationHistorySegmentResolver {
	return res.LocationHistorySegment()
}
