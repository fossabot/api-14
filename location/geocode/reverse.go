package geocode

import (
	"go.stevenxie.me/api/v2/location"
)

type (
	// ReverseGeocodeOptions are option parameters for a reverse-geocoding
	// request.
	ReverseGeocodeOptions struct {
		Level           Level
		Radius          uint
		IncludeShape    bool
		IncludeTimeZone bool
	}

	// A ReverseGeocodeOption modifies a ReverseGeocodeOptions.
	ReverseGeocodeOption func(*ReverseGeocodeOptions)

	// A ReverseGeocodeResult is the result of a reverse-geocoding search.
	ReverseGeocodeResult struct {
		Place     location.Place
		Relevance float32
		Distance  float32
	}
)

// ReverseWithLevel sets the geocoding level (proximity) of a
// reverse-geocoding request.
func ReverseWithLevel(l Level) ReverseGeocodeOption {
	return func(opt *ReverseGeocodeOptions) { opt.Level = l }
}

// ReverseWithRadius sets the search radius of a reverse-geocoding
// request.
func ReverseWithRadius(radius uint) ReverseGeocodeOption {
	return func(opt *ReverseGeocodeOptions) { opt.Radius = radius }
}

// ReverseWithShape sets a reverse-geocoding request to include
// geographical shape information in the response.
func ReverseWithShape(include bool) ReverseGeocodeOption {
	return func(opt *ReverseGeocodeOptions) { opt.IncludeShape = include }
}

// ReverseWithTimeZone sets a reverse-geocoding request to include
// time zone information in the response.
func ReverseWithTimeZone(include bool) ReverseGeocodeOption {
	return func(opt *ReverseGeocodeOptions) { opt.IncludeTimeZone = include }
}
