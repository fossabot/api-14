package schedsvc

import (
	"context"
	"sort"
	"time"

	"github.com/cockroachdb/errors"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"

	"go.stevenxie.me/gopkg/logutil"
	"go.stevenxie.me/gopkg/name"

	"go.stevenxie.me/api/v2/location"
	"go.stevenxie.me/api/v2/pkg/basic"
	"go.stevenxie.me/api/v2/scheduling"
)

// NewService creates a new Service.
func NewService(
	cal scheduling.Calendar,
	zones location.TimeZoneService,
	opts ...basic.Option,
) scheduling.Service {
	cfg := basic.BuildOptions(opts...)
	return service{
		cal:    cal,
		zones:  zones,
		log:    logutil.WithComponent(cfg.Logger, (*service)(nil)),
		tracer: cfg.Tracer,
	}
}

type service struct {
	cal   scheduling.Calendar
	zones location.TimeZoneService

	log    *logrus.Entry
	tracer opentracing.Tracer
}

var _ scheduling.Service = (*service)(nil)

func (svc service) BusyTimes(
	ctx context.Context,
	date time.Time,
) ([]scheduling.TimeSpan, error) {
	span, ctx := opentracing.StartSpanFromContextWithTracer(
		ctx, svc.tracer,
		name.OfFunc(service.BusyTimes),
	)
	defer span.Finish()

	log := svc.log.WithFields(logrus.Fields{
		logutil.MethodKey: name.OfMethod(service.BusyTimes),
		"date":            date,
	}).WithContext(ctx)

	log.Trace("Getting busy times from calendar...")
	periods, err := svc.cal.BusyTimes(ctx, date)
	if err != nil {
		log.WithError(err).Error("Failed to load busy times from calendar.")
		return nil, err
	}
	log.
		WithField("periods", periods).
		Trace("Loaded busy times from calendar.")

	// Sort periods.
	sort.Slice(periods, func(i, j int) bool {
		return periods[i].Before(&periods[j])
	})
	log.
		WithField("periods", periods).
		Trace("Sorted busy times by time.")
	return periods, nil
}

func (svc service) BusyTimesToday(ctx context.Context) ([]scheduling.TimeSpan, error) {
	span, ctx := opentracing.StartSpanFromContextWithTracer(
		ctx, svc.tracer,
		name.OfFunc(service.BusyTimesToday),
	)
	defer span.Finish()

	log := logutil.
		WithMethod(svc.log, service.BusyTimesToday).
		WithContext(ctx)

	log.Trace("Getting current time zone...")
	tz, err := svc.zones.CurrentTimeZone(ctx)
	if err != nil {
		log.WithError(err).Error("Failed to get current time zone.")
		return nil, errors.Wrap(err, "schedsvc: get current time zone")
	}
	return svc.BusyTimes(ctx, time.Now().In(tz))
}
