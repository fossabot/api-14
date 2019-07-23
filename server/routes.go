package server

import (
	"github.com/sirupsen/logrus"
	"go.stevenxie.me/api/server/handler"
	"go.stevenxie.me/api/service/music"
)

func (srv *Server) registerRoutes() error {
	e := srv.echo
	e.HTTPErrorHandler = handler.ErrorHandler(srv.hlog("error"))

	// Register routes.
	e.GET("/", handler.InfoHandler())
	e.GET("/about", handler.AboutHandler(srv.about, srv.hlog("about")))
	e.GET(
		"/productivity",
		handler.ProductivityHandler(srv.productivity, srv.hlog("productivity")),
	)
	e.GET(
		"/availability",
		handler.AvailabilityHandler(srv.availability, srv.hlog("availability")),
	)
	e.GET(
		"/commits",
		handler.RecentCommitsHandler(srv.commits, srv.hlog("recent_commits")),
	)

	// Register location routes.
	location := handler.NewLocationProvider(srv.location)
	e.GET(
		"/location",
		location.CurrentRegionHandler(srv.hlog("location_currrent_region")),
	)
	e.GET(
		"/location/history",
		location.HistoryHandler(
			srv.locationAccess,
			srv.hlog("location_history"),
		),
	)

	// Handle music routes.
	e.GET(
		"/nowplaying",
		handler.NowPlayingHandler(srv.nowplaying, srv.hlog("nowplaying")),
	)
	if streamer, ok := srv.nowplaying.(music.NowPlayingStreamingService); ok {
		e.GET(
			"/nowplaying/ws",
			handler.NowPlayingStreamingHandler(
				streamer,
				srv.hlog("nowplaying_streaming"),
			),
		)
	} else {
		srv.log.Warn("No music streaming service available; nowplaying streams " +
			"are disabled.")
	}

	return nil
}

func (srv *Server) hlog(name string) *logrus.Logger {
	return srv.log.WithField("handler", name).Logger
}
