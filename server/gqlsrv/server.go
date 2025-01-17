package gqlsrv

import (
	"context"
	"io/ioutil"
	"net/http"

	"go.stevenxie.me/gopkg/configutil"

	"github.com/cockroachdb/errors"
	sentry "github.com/getsentry/sentry-go"
	"github.com/sirupsen/logrus"
	"go.stevenxie.me/gopkg/logutil"

	echo "github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"go.stevenxie.me/api/v2/about"
	"go.stevenxie.me/api/v2/assist/transit"
	"go.stevenxie.me/api/v2/auth"
	"go.stevenxie.me/api/v2/git"
	"go.stevenxie.me/api/v2/location"
	"go.stevenxie.me/api/v2/music"
	"go.stevenxie.me/api/v2/productivity"
	"go.stevenxie.me/api/v2/scheduling"
)

// NewServer creates a new Server.
func NewServer(svcs Services, strms Streamers, opts ...ServerOption) *Server {
	cfg := ServerOptions{
		Logger:          logutil.NoopEntry(),
		ComplexityLimit: 5,
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	// Configure logger.
	log := logutil.WithComponent(cfg.Logger, (*Server)(nil))

	// Configure Echo.
	echo := echo.New()
	echo.Logger.SetOutput(ioutil.Discard) // disable logger

	// Configure middleware.
	echo.Pre(middleware.RemoveTrailingSlashWithConfig(
		middleware.TrailingSlashConfig{
			RedirectCode: http.StatusPermanentRedirect,
		},
	))
	echo.Use(middleware.Recover())

	// Enable Access-Control-Allow-Origin: * during development.
	if configutil.GetGoEnv() == configutil.GoEnvDevelopment {
		echo.Use(middleware.CORS())
	}

	// Create server.
	return &Server{
		echo:   echo,
		log:    log,
		sentry: cfg.Sentry,

		svcs:  svcs,
		strms: strms,

		complexityLimit: cfg.ComplexityLimit,
	}
}

// WithLogger configures a Server to write logs with log.
func WithLogger(log *logrus.Entry) ServerOption {
	return func(opt *ServerOptions) { opt.Logger = log }
}

// WithSentry configures a server to capture handler panics with hub.
func WithSentry(hub *sentry.Hub) ServerOption {
	return func(opt *ServerOptions) { opt.Sentry = hub }
}

// WithComplexityLimit configures a Server to limit GraphQL queries by
// complexity.
func WithComplexityLimit(limit int) ServerOption {
	return func(opt *ServerOptions) { opt.ComplexityLimit = limit }
}

type (
	// Server serves the accounts REST API.
	Server struct {
		echo   *echo.Echo
		log    *logrus.Entry
		sentry *sentry.Hub

		svcs  Services
		strms Streamers

		complexityLimit int
	}

	// Services are used to handle server requests.
	Services struct {
		Git          git.Service
		Auth         auth.Service
		About        about.Service
		Music        music.Service
		Transit      transit.Service
		Location     location.Service
		Scheduling   scheduling.Service
		Productivity productivity.Service
	}

	// Streamers are used to handle server streams.
	Streamers struct {
		Music music.Streamer
	}

	// A ServerOptions configures a Server.
	ServerOptions struct {
		Logger *logrus.Entry
		Sentry *sentry.Hub

		// Complexity limit for GraphQL queries.
		ComplexityLimit int
	}

	// An ServerOption modifies a ServerOptions.
	ServerOption func(*ServerOptions)
)

// ListenAndServe listens and serves on the specified address.
func (srv *Server) ListenAndServe(addr string) error {
	if addr == "" {
		return errors.New("gqlsrv: addr must be non-empty")
	}
	log := srv.log.WithField("addr", addr)

	// Register routes.
	if err := srv.registerRoutes(); err != nil {
		return errors.Wrap(err, "gqlsrv: registering routes")
	}

	// Listen for connections.
	log.Info("Listening for connections...")
	return srv.echo.Start(addr)
}

// Shutdown shuts down the server gracefully without interupting any active
// connections.
func (srv *Server) Shutdown(ctx context.Context) error {
	return srv.echo.Shutdown(ctx)
}
