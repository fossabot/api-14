package stream

import (
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	errors "golang.org/x/xerrors"

	"github.com/stevenxie/api/pkg/geo"
	"github.com/stevenxie/api/pkg/zero"
)

type (
	// A RecentLocationsPreloader implements a geo.RecentLocationsService that
	// preloads recent locations data.
	RecentLocationsPreloader struct {
		streamer *PollStreamer
		geocoder geo.Geocoder
		log      *logrus.Logger

		mux     sync.Mutex
		segment *geo.Segment
		err     error
	}

	// An LSOption configures a LocationPreloader.
	LSOption func(*RecentLocationsPreloader)
)

var _ geo.RecentLocationsService = (*RecentLocationsPreloader)(nil)

// NewRecentLocationsPreloader creates a new LocationPreloader.
func NewRecentLocationsPreloader(
	locations geo.RecentLocationsService,
	geo geo.Geocoder,
	interval time.Duration,
	opts ...LSOption,
) *RecentLocationsPreloader {
	ls := &RecentLocationsPreloader{
		geocoder: geo,
		log:      zero.Logger(),
	}
	for _, opt := range opts {
		opt(ls)
	}

	// Configure streamer.
	action := func() (zero.Interface, error) { return locations.LastSegment() }
	ls.streamer = NewPollStreamer(action, interval)

	go ls.populateCache()
	return ls
}

// WithLSLogger configures a LocationPreloader's logger.
func WithLSLogger(log *logrus.Logger) LSOption {
	return func(lp *RecentLocationsPreloader) { lp.log = log }
}

func (ls *RecentLocationsPreloader) populateCache() {
	for result := range ls.streamer.Stream() {
		var (
			segment *geo.Segment
			err     error
		)

		switch v := result.(type) {
		case error:
			err = v
			ls.log.WithError(err).Error("Failed to load last seen position.")
		case *geo.Segment:
			segment = v
		default:
			ls.log.WithField("value", v).Error("Unexpected value from upstream.")
			err = errors.Errorf("stream: unexpected value '%s' from upstream")
		}

		ls.mux.Lock()
		ls.segment = segment
		ls.err = err
		ls.mux.Unlock()
	}
}

// Stop stops the LocationPreloader.
func (ls *RecentLocationsPreloader) Stop() { ls.streamer.Stop() }

// LastSegment returns the authenticated user's latest location history segment.
func (ls *RecentLocationsPreloader) LastSegment() (*geo.Segment, error) {
	ls.mux.Lock()
	defer ls.mux.Unlock()
	copy := *ls.segment
	return &copy, nil
}