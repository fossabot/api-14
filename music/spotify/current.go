package spotify

import (
	"context"
	"io"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/sirupsen/logrus"
	"github.com/zmb3/spotify"

	"go.stevenxie.me/api/music"
	"go.stevenxie.me/api/pkg/svcutil"
	"go.stevenxie.me/gopkg/logutil"
)

// NewCurrentService creates a new music.CurrentSource.
func NewCurrentService(
	c *spotify.Client,
	opts ...svcutil.BasicOption,
) music.CurrentService {
	cfg := svcutil.BasicConfig{
		Logger: logutil.NoopEntry(),
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return currentService{
		client: c,
		log:    logutil.AddComponent(cfg.Logger, (*currentService)(nil)),
	}
}

type currentService struct {
	client *spotify.Client
	log    *logrus.Entry
}

var _ music.CurrentService = (*currentService)(nil)

func (svc currentService) GetCurrent(ctx context.Context) (*music.CurrentlyPlaying,
	error) {
	log := logutil.
		WithMethod(svc.log, currentService.GetCurrent).
		WithContext(ctx)

	log.Trace("Requesting currently playing info from Spotify.")
	cp, err := svc.client.PlayerCurrentlyPlaying()
	if err != nil {
		if errors.Is(err, io.EOF) { // nothing is playing
			log.Trace("Nothing is playing right now.")
			return nil, nil
		}
		log.WithError(err).Error("Failed to get currently playing info from Spotify.")
		return nil, err
	}
	log.WithField("current", cp).Trace("Got currently playing info.")

	item := cp.Item
	if item == nil {
		return nil, nil
	}

	// Parse timestamp as time.Time.
	timestamp := time.Unix(
		cp.Timestamp/1000,
		(cp.Timestamp%1000)*int64((time.Millisecond/time.Nanosecond)),
	)
	log.
		WithField("timestamp", timestamp).
		Trace("Parsed unix timestamp as a time.Time.")

	// Parse track.
	track := music.Track{Album: new(music.Album)}
	trackFromSpotify(&track, &item.SimpleTrack)
	albumFromSpotify(track.Album, &item.Album)
	log.
		WithField("track", track).
		Trace("Marshal response to music.Track.")

	return &music.CurrentlyPlaying{
		Timestamp: timestamp,
		Playing:   cp.Playing,
		Progress:  time.Duration(cp.Progress) * time.Millisecond,
		Track:     track,
	}, nil
}
