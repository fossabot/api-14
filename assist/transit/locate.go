package transit

import (
	"context"

	validation "github.com/go-ozzo/ozzo-validation"
	"go.stevenxie.me/api/v2/location"
)

type (
	// A Locator can locate nearby departures.
	Locator interface {
		NearbyDepartures(
			ctx context.Context,
			coords location.Coordinates,
			opt NearbyDeparturesOptions,
		) ([]NearbyDeparture, error)
	}

	// A NearbyDeparture is a Departure that is occurring nearby.
	NearbyDeparture struct {
		Departure `json:"departure"`

		// How far away the departure is, in meters.
		Distance int `json:"distance"`
	}
)

type (
	// A LocatorService wraps a Locator with a friendlier API and logging.
	//
	// Results are sorted in ascending order by distance.
	LocatorService interface {
		NearbyDepartures(
			ctx context.Context,
			pos location.Coordinates,
			opts ...NearbyDeparturesOption,
		) ([]NearbyDeparture, error)
	}

	// NearbyDeparturesOptions are option parameters for Locator.NearbyDepartures.
	NearbyDeparturesOptions struct {
		Radius          int // the search radius, in meters
		MaxStations     int // max number of stations to look up
		MaxPerStation   int // max number of departures per station
		MaxPerTransport int // max departures per transport
	}

	// A NearbyDeparturesOption modifies a NearbyDepartureOptions.
	NearbyDeparturesOption func(*NearbyDeparturesOptions)
)

var _ validation.Validatable = (*NearbyDeparturesOptions)(nil)

// Validate returns an error if the config is not valid.
func (cfg *NearbyDeparturesOptions) Validate() error {
	nonNegFields := []*int{&cfg.Radius, &cfg.MaxStations, &cfg.MaxPerStation}
	rules := make([]*validation.FieldRules, len(nonNegFields))
	for i, f := range nonNegFields {
		rules[i] = validation.Field(f, validation.Min(0))
	}
	return validation.ValidateStruct(cfg, rules...)
}
