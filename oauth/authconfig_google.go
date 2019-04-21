package oauth

import (
	"github.com/gorilla/sessions"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/Peanoquio/oauthserver/common"
)

// AuthConfigGoogle struct
type AuthConfigGoogle struct {
	envConfig    *common.EnvVars
	OAuthConfig  *oauth2.Config
	SessionStore sessions.Store
}

// init will initialize the config
func (config *AuthConfigGoogle) init(envConfig *common.EnvVars) {
	config.envConfig = envConfig

	// configure the Google Oauth2 client
	clientID := envConfig.GoogleOauth2ClientID
	clientSecret := envConfig.GoogleOauth2ClientSecret
	config.OAuthConfig = config.configureOAuthClient(clientID, clientSecret)

	// configure the cookie that will store the session information after login
	cookieStoreSecret := envConfig.CookieStoreSecret
	cookieStore := sessions.NewCookieStore([]byte(cookieStoreSecret))
	cookieStore.Options = &sessions.Options{
		HttpOnly: true,
	}
	config.SessionStore = cookieStore
}

// configureGoogleOAuthClient configures the Google oauth client
// https://console.cloud.google.com/apis/credentials
func (config *AuthConfigGoogle) configureOAuthClient(clientID, clientSecret string) *oauth2.Config {
	redirectURL := config.envConfig.GoogleOauth2Callback

	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{"email", "profile"},
		Endpoint:     google.Endpoint,
	}
}
