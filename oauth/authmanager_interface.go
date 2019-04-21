package oauth

import (
	"net/http"

	"github.com/Peanoquio/oauthserver/common"
)

// AuthManagerInterface serves the interface that needs to be implemented
type AuthManagerInterface interface {
	Init(envConfig *common.EnvVars)
	IsAuthenticated(r *http.Request) bool
	OauthPageHandler(w http.ResponseWriter, r *http.Request) *common.AppError
	LoginHandler(w http.ResponseWriter, r *http.Request) *common.AppError
	LogoutHandler(w http.ResponseWriter, r *http.Request) *common.AppError
	OauthCallbackHandler(w http.ResponseWriter, r *http.Request) *common.AppError
}
