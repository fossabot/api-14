package gqlutil

import (
	"context"
	"net/http"

	"github.com/cockroachdb/errors/exthttp"

	"github.com/99designs/gqlgen/graphql"
	"github.com/cockroachdb/errors"
	"github.com/vektah/gqlparser/gqlerror"

	"go.stevenxie.me/gopkg/zero"
)

// PresentError is a graphql.ErrorPresenterFunc that knows how to present
// errors created by cockroachdb/errors.
func PresentError(ctx context.Context, err error) *gqlerror.Error {
	exts := make(map[string]zero.Interface)

	// Append to extensions.
	if hints := errors.GetAllHints(err); len(hints) > 0 {
		exts["hints"] = hints
	}
	if details := errors.GetAllDetails(err); len(details) > 0 {
		exts["details"] = details
	}
	if links := errors.GetAllIssueLinks(err); len(links) > 0 {
		exts["issueLinks"] = links
	}
	if cause := errors.UnwrapAll(err); (cause != err) && (cause != nil) {
		if msg := cause.Error(); msg != err.Error() {
			exts["cause"] = msg
		}
	}
	if trace := errors.GetReportableStackTrace(err); trace != nil {
		exts["culprit"] = trace.Culprit()
	}
	if file, line, fn, ok := errors.GetOneLineSource(err); ok {
		exts["source"] = struct {
			File string `json:"file"`
			Line int    `json:"line"`
			Fn   string `json:"fn"`
		}{file, line, fn}
	}
	{
		status := exthttp.GetHTTPCode(err, http.StatusInternalServerError)
		exts["status"] = struct {
			Code int    `json:"code"`
			Text string `json:"text"`
		}{status, http.StatusText(status)}
	}

	gqlErr := graphql.DefaultErrorPresenter(ctx, err)
	if len(exts) > 0 {
		if len(gqlErr.Extensions) == 0 {
			gqlErr.Extensions = exts
		} else {
			for k, v := range exts {
				gqlErr.Extensions[k] = v
			}
		}
	}
	return gqlErr
}

var _ graphql.ErrorPresenterFunc = PresentError
