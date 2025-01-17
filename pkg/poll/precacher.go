package poll

import (
	"errors"
	"time"

	"github.com/sirupsen/logrus"
	"go.stevenxie.me/gopkg/logutil"
	"go.stevenxie.me/gopkg/zero"
)

// NewPrecacher creates a new Precacher that caches the values produced by a Producer.
//
// It produces new values at an interval of n.
func NewPrecacher(
	p Producer,
	n time.Duration,
	opts ...PrecacherOption,
) *Precacher {
	opt := PrecacherOptions{
		Logger: logutil.NoopEntry(),
	}
	for _, apply := range opts {
		apply(&opt)
	}

	var (
		log = logutil.WithComponent(opt.Logger, (*Precacher)(nil))
		ca  = newCacheActor(p, log)
	)
	return &Precacher{
		pl: NewPoller(
			ca, n,
			PollerWithLogger(log),
		),
		ca: ca,
	}
}

// PrecacherWithLogger configures a Precacher to write logs with log.
func PrecacherWithLogger(log *logrus.Entry) PrecacherOption {
	return func(opt *PrecacherOptions) { opt.Logger = log }
}

type (
	// A Precacher caches the values produced by a Producer at regular intervals.
	Precacher struct {
		pl *Poller
		ca *cacheActor
	}

	// A PrecacherOptions configures a Precacher.
	PrecacherOptions struct {
		Logger *logrus.Entry
	}

	// A PrecacherOption modifies a PrecacherOptions.
	PrecacherOption func(*PrecacherOptions)
)

// Results returns the latest cached results.
func (pc Precacher) Results() (val zero.Interface, err error) {
	return pc.ca.Results()
}

// Stop stops the Precacher from requesting new values from its Producer.
func (pc Precacher) Stop() { pc.pl.Stop() }

func newCacheActor(p Producer, log *logrus.Entry) *cacheActor {
	return &cacheActor{
		Producer: p,
		log:      log,
		empty:    true,
	}
}

type cacheActor struct {
	Producer
	result
	log   *logrus.Entry
	empty bool
}

func (ca *cacheActor) Recv(v zero.Interface, err error) {
	log := ca.log.WithField("value", v)
	ca.empty = false
	ca.Error = err

	// Only save new value if it is non-nil; that way, the previous value
	// can be used in the event that an error occurs.
	if v != nil {
		ca.Value = v
	}

	if err != nil {
		log.
			WithError(err).
			Error("Received error from Producer.")
	} else {
		log.Trace("Received value from Producer.")
	}
}

func (ca *cacheActor) Results() (zero.Interface, error) {
	if ca.empty {
		return nil, ErrCacheEmpty
	}
	res := ca.result
	return res.Value, res.Error
}

// ErrCacheEmpty is returned by Precacher.Results when no values have been
// received yet from the Producer.
var ErrCacheEmpty = errors.New("poll: empty cache")
