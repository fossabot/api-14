package spotify

import (
	"os"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
)

// Namespace is the package namespace, used for things like envvars.
const Namespace = "spotify"

// New creates a new Spotify client.
//
// It reads SPOTIFY_TOKEN (the refresh token) from the environment; if no such
// variable is found, an error will be returned.
func New() (*spotify.Client, error) {
	refresh := os.Getenv(strings.ToUpper(Namespace) + "_TOKEN")
	if refresh == "" {
		return nil, errors.Newf(
			"spotify: no such environment variable '%s_TOKEN'",
			strings.ToUpper(Namespace),
		)
	}

	var (
		token  = &oauth2.Token{RefreshToken: refresh, TokenType: "Bearer"}
		client = spotify.NewAuthenticator("", "").NewClient(token)
	)
	return &client, nil
}
