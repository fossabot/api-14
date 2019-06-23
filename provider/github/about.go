package github

import (
	"encoding/json"
	"time"

	"github.com/stevenxie/api/pkg/api"
	errors "golang.org/x/xerrors"
)

// AboutService reads an about.Info from a file stored in Github Gists.
type AboutService struct {
	repo         GistRepo
	gistID, file string
}

// A GistRepo can retrieve gist data.
type GistRepo interface {
	GistFile(id, file string) ([]byte, error)
}

var _ api.AboutService = (*AboutService)(nil)

// NewAboutService creates a new AboutService that reads Info from a GitHub
// gist.
func NewAboutService(gr GistRepo, gistID, file string) *AboutService {
	return &AboutService{
		repo:   gr,
		gistID: gistID,
		file:   file,
	}
}

// About retrieves About info from a GitHub gist.
func (as *AboutService) About() (*api.About, error) {
	raw, err := as.repo.GistFile(as.gistID, as.file)
	if err != nil {
		return nil, errors.Errorf("github: getting gist: %w", err)
	}

	// Decode gist contents.
	var data struct {
		*api.About
		Birthday string `json:"birthday"`
	}
	if err = json.Unmarshal(raw, &data); err != nil {
		return nil, errors.Errorf("github: decoding gist file contents as JSON: %w",
			err)
	}

	// Derive age from birthday.
	bday, err := time.Parse("2006-01-02", data.Birthday)
	if err != nil {
		return nil, errors.Errorf("github: failed to parse birthday '%s': %w",
			data.Birthday, err)
	}
	data.About.Age = time.Since(bday).Truncate(365 * 24 * time.Hour)

	// Fill missing values.
	if data.About.Whereabouts == "" {
		data.About.Whereabouts = "NOT IMPLEMENTED"
	}

	return data.About, nil
}