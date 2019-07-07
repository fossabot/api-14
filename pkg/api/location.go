package api

import (
	"github.com/stevenxie/api/pkg/geo"
)

type (
	// A LocationService provides information about my recent locations.
	LocationService interface {
		CurrentCity() (city string, err error)
		CurrentRegion() (*geo.Location, error)
		LastPosition() (*geo.Coordinate, error)
		LastSegment() (*geo.Segment, error)
	}

	// A LocationAccessService can validate location access codes.
	LocationAccessService interface {
		IsValidCode(code string) (valid bool, err error)
	}
)