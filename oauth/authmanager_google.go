package oauth

import (
	"context"
	"encoding/gob"
	"errors"
	"net/http"
	"net/url"

	"github.com/Peanoquio/oauthserver/common"
	"github.com/Peanoquio/oauthserver/view"

	uuid "github.com/gofrs/uuid"
	"golang.org/x/oauth2"
	plus "google.golang.org/api/plus/v1"
)

const (
	defaultSessionID = "default"
	// keys for the SessionStore
	googleProfileSessionKey = "google_profile"
	oauthTokenSessionKey    = "oauth_token"

	// This key is used in the OAuth flow session to store the URL to redirect the
	// user to after the OAuth flow is complete.
	oauthFlowRedirectKey = "redirect"

	// the time-to-live of the session in seconds
	sessionTTLSecs = 10 * 60
)

// init initializes this auth manager
func init() {
	// Gob encoding for gorilla/sessions
	gob.Register(&oauth2.Token{})
	gob.Register(&common.Profile{})
}

// AuthManagerGoogle handles Google related authentication
type AuthManagerGoogle struct {
	authConfig *AuthConfigGoogle
	templates  map[string]*view.AppTemplate
}

// Init will initialize the AuthManagerGoogle struct
func (authMgr *AuthManagerGoogle) Init(envConfig *common.EnvVars) {
	// init the Google config
	authMgr.authConfig = &AuthConfigGoogle{}
	authMgr.authConfig.init(envConfig)

	// initialize and set the template pages
	authMgr.templates = make(map[string]*view.AppTemplate)
	authMgr.templates["oauth"] = view.ParseTemplate("oauth.html")
}

// IsAuthenticated checks if the user has been authenticated based on the request
func (authMgr *AuthManagerGoogle) IsAuthenticated(r *http.Request) bool {
	profile := authMgr.profileFromSession(r)
	if profile != nil {
		return true
	}
	return false
}

// OauthPageHandler displays the default oauth login page
func (authMgr *AuthManagerGoogle) OauthPageHandler(w http.ResponseWriter, r *http.Request) *common.AppError {
	isAuthEnabled := authMgr.authConfig != nil
	profile := authMgr.profileFromSession(r)
	var data interface{}
	return authMgr.templates["oauth"].Execute(w, r, isAuthEnabled, profile, data)
}

// LoginHandler initiates an OAuth flow to authenticate the user.
func (authMgr *AuthManagerGoogle) LoginHandler(w http.ResponseWriter, r *http.Request) *common.AppError {
	sessionID := uuid.Must(uuid.NewV4()).String()
	// create the session based on the uniquely generated session ID
	oauthFlowSession, err := authMgr.authConfig.SessionStore.New(r, sessionID)
	if err != nil {
		return common.AppErrorf(err, "could not create oauth session: %v", err)
	}
	// session duration before expiry
	oauthFlowSession.Options.MaxAge = sessionTTLSecs

	redirectURL, err := authMgr.validateRedirectURL(r.FormValue("redirect"))
	if err != nil {
		return common.AppErrorf(err, "invalid redirect URL: %v", err)
	}
	oauthFlowSession.Values[oauthFlowRedirectKey] = redirectURL

	// save the session in the response
	if err := oauthFlowSession.Save(r, w); err != nil {
		return common.AppErrorf(err, "could not save session: %v", err)
	}

	// Use the session ID for the "state" parameter.
	// This protects against CSRF (cross-site request forgery).
	// See https://godoc.org/golang.org/x/oauth2#AuthConfig.AuthCodeURL for more detail.
	url := authMgr.authConfig.OAuthConfig.AuthCodeURL(sessionID, oauth2.ApprovalForce, oauth2.AccessTypeOnline)
	http.Redirect(w, r, url, http.StatusFound)

	return nil
}

// validateRedirectURL checks that the URL provided is valid.
// If the URL is missing, redirect the user to the application's root.
// The URL must not be absolute (i.e., the URL must refer to a path within this application).
func (authMgr *AuthManagerGoogle) validateRedirectURL(path string) (string, error) {
	if path == "" {
		return "/", nil
	}

	// Ensure redirect URL is valid and not pointing to a different server.
	parsedURL, err := url.Parse(path)
	if err != nil {
		return "/", err
	}
	if parsedURL.IsAbs() {
		return "/", errors.New("URL must not be absolute")
	}
	return path, nil
}

// OauthCallbackHandler completes the OAuth flow, retreives the user's profile information and stores it in a session.
// This is the callback that will be called by Google after the user completes authentication
func (authMgr *AuthManagerGoogle) OauthCallbackHandler(w http.ResponseWriter, r *http.Request) *common.AppError {
	// The Google authorization service sends the user back to your application via the oauth callback URL path
	// along with the authorization code specified by the code form value
	oauthFlowSession, err := authMgr.authConfig.SessionStore.Get(r, r.FormValue("state"))
	if err != nil {
		return common.AppErrorf(err, "invalid state parameter. try logging in again.")
	}

	redirectURL, ok := oauthFlowSession.Values[oauthFlowRedirectKey].(string)
	// Validate this callback request came from the app.
	if !ok {
		return common.AppErrorf(err, "invalid state parameter. try logging in again.")
	}

	// Exchange converts an authorization code into a token.
	code := r.FormValue("code")
	token, err := authMgr.authConfig.OAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		return common.AppErrorf(err, "could not get auth token: %v", err)
	}

	session, err := authMgr.authConfig.SessionStore.New(r, defaultSessionID)
	if err != nil {
		return common.AppErrorf(err, "could not get default session: %v", err)
	}

	ctx := context.Background()
	// get the user Google profile through the token
	profile, err := authMgr.fetchProfile(ctx, token)
	if err != nil {
		return common.AppErrorf(err, "could not fetch Google profile: %v", err)
	}

	// store the token and Google profile in the session
	session.Values[oauthTokenSessionKey] = token
	session.Values[googleProfileSessionKey] = stripProfile(profile)
	if err := session.Save(r, w); err != nil {
		return common.AppErrorf(err, "could not save session: %v", err)
	}

	http.Redirect(w, r, redirectURL, http.StatusFound)
	return nil
}

// fetchProfile retrieves the Google+ profile of the user associated with the provided OAuth token.
func (authMgr *AuthManagerGoogle) fetchProfile(ctx context.Context, token *oauth2.Token) (*plus.Person, error) {
	client := oauth2.NewClient(ctx, authMgr.authConfig.OAuthConfig.TokenSource(ctx, token))
	plusService, err := plus.New(client)
	if err != nil {
		return nil, err
	}
	return plusService.People.Get("me").Do()
}

// LogoutHandler clears the default session.
func (authMgr *AuthManagerGoogle) LogoutHandler(w http.ResponseWriter, r *http.Request) *common.AppError {
	session, err := authMgr.authConfig.SessionStore.New(r, defaultSessionID)
	if err != nil {
		return common.AppErrorf(err, "could not get default session: %v", err)
	}
	session.Options.MaxAge = -1 // Clear session.
	if err := session.Save(r, w); err != nil {
		return common.AppErrorf(err, "could not save session: %v", err)
	}
	redirectURL := r.FormValue("redirect")
	if redirectURL == "" {
		redirectURL = "/"
	}
	http.Redirect(w, r, redirectURL, http.StatusFound)
	return nil
}

// profileFromSession retreives the Google+ profile from the default session.
// Returns nil if the profile cannot be retreived (e.g. user is logged out).
func (authMgr *AuthManagerGoogle) profileFromSession(r *http.Request) *common.Profile {
	session, err := authMgr.authConfig.SessionStore.Get(r, defaultSessionID)
	if err != nil {
		return nil
	}
	tok, ok := session.Values[oauthTokenSessionKey].(*oauth2.Token)
	if !ok || !tok.Valid() {
		return nil
	}
	profile, ok := session.Values[googleProfileSessionKey].(*common.Profile)
	if !ok {
		return nil
	}
	return profile
}

// stripProfile returns a subset of a plus.Person.
func stripProfile(p *plus.Person) *common.Profile {
	return &common.Profile{
		ID:          p.Id,
		DisplayName: p.DisplayName,
		ImageURL:    p.Image.Url,
	}
}
