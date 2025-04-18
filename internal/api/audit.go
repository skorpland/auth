package api

import (
	"net/http"
	"strings"

	"github.com/skorpland/auth/internal/api/apierrors"
	"github.com/skorpland/auth/internal/models"
)

var filterColumnMap = map[string][]string{
	"author": {"actor_username", "actor_name"},
	"action": {"action"},
	"type":   {"log_type"},
}

func (a *API) adminAuditLog(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	db := a.db.WithContext(ctx)

	// aud := a.requestAud(ctx, r)
	pageParams, err := paginate(r)
	if err != nil {
		return badRequestError(apierrors.ErrorCodeValidationFailed, "Bad Pagination Parameters: %v", err)
	}

	var col []string
	var qval string
	q := r.URL.Query().Get("query")
	if q != "" {
		var exists bool
		qparts := strings.SplitN(q, ":", 2)
		col, exists = filterColumnMap[qparts[0]]
		if !exists || len(qparts) < 2 {
			return badRequestError(apierrors.ErrorCodeValidationFailed, "Invalid query scope: %s", q)
		}
		qval = qparts[1]
	}

	logs, err := models.FindAuditLogEntries(db, col, qval, pageParams)
	if err != nil {
		return internalServerError("Error searching for audit logs").WithInternalError(err)
	}

	addPaginationHeaders(w, r, pageParams)

	return sendJSON(w, http.StatusOK, logs)
}
