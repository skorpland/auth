package api

import (
	"net/http"

	"github.com/skorpland/auth/internal/api/apierrors"
	"github.com/skorpland/auth/internal/metering"
	"github.com/skorpland/auth/internal/models"
	"github.com/skorpland/auth/internal/storage"
)

func (a *API) SignupAnonymously(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	config := a.config
	db := a.db.WithContext(ctx)
	aud := a.requestAud(ctx, r)

	if config.DisableSignup {
		return unprocessableEntityError(apierrors.ErrorCodeSignupDisabled, "Signups not allowed for this instance")
	}

	params := &SignupParams{}
	if err := retrieveRequestParams(r, params); err != nil {
		return err
	}
	params.Aud = aud
	params.Provider = "anonymous"

	newUser, err := params.ToUserModel(false /* <- isSSOUser */)
	if err != nil {
		return err
	}

	var grantParams models.GrantParams
	grantParams.FillGrantParams(r)

	var token *AccessTokenResponse
	err = db.Transaction(func(tx *storage.Connection) error {
		var terr error
		newUser, terr = a.signupNewUser(tx, newUser)
		if terr != nil {
			return terr
		}
		token, terr = a.issueRefreshToken(r, tx, newUser, models.Anonymous, grantParams)
		if terr != nil {
			return terr
		}
		return nil
	})
	if err != nil {
		return internalServerError("Database error creating anonymous user").WithInternalError(err)
	}

	metering.RecordLogin("anonymous", newUser.ID)
	return sendJSON(w, http.StatusOK, token)
}
