package handlers

import (
	"fmt"
	"net/http"

	"github.com/sogelink-research/pgrest/errors"
)

func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	details := fmt.Sprintf("Path '%s' not found", r.URL.Path)
	error := errors.NewAPIError(http.StatusNotFound, "Not found", &details)
	HandleError(w, error)
}
