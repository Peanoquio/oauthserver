package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Peanoquio/oauthserver/common"
	"github.com/Peanoquio/oauthserver/enums"
	"github.com/Peanoquio/oauthserver/oauth"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

// main executable to start the server
func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	envConfig := common.LoadEnvironmentVariables()
	registerHandlers(envConfig)

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}

// registerHandler maps the routing path to the request handler based on the platform
// TODO Oliver to refactor the routing based on the platform of the authMgr (Google, Facebook, etc.)
func registerHandler(envConfig *common.EnvVars, platform enums.Platform, router *mux.Router) {
	authMgr := oauth.NewAuthManager(enums.Google, envConfig)

	// When registering handlers with the http package we use the Handle function (instead of HandleFunc)
	// as appHandler is an http.Handler (not an http.HandlerFunc)
	router.Methods("GET").Path("/oauthpage").Handler(appHandler(authMgr.OauthPageHandler))
	router.Methods("POST").Path("/login").Handler(appHandler(authMgr.LoginHandler))
	router.Methods("POST").Path("/logout").Handler(appHandler(authMgr.LogoutHandler))
	router.Methods("GET").Path("/oauth2callback").Handler(appHandler(authMgr.OauthCallbackHandler))

	// test if the user has been authenticated
	router.Methods("GET").Path("/test").HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			isAuthenticated := authMgr.IsAuthenticated(r)
			if isAuthenticated == false {
				http.Error(w, "Forbidden", http.StatusForbidden)
			} else {
				w.Write([]byte("test is successful since you have been authenticated"))
			}
		})
}

// registerHandlers will manage the routing paths and map it to the request handlers
func registerHandlers(envConfig *common.EnvVars) {
	// Use gorilla/mux for rich routing.
	// See http://www.gorillatoolkit.org/pkg/mux
	router := mux.NewRouter()

	registerHandler(envConfig, enums.Google, router)

	// redirect to the test path endpoint
	router.Handle("/", http.RedirectHandler("/test", http.StatusFound))

	// health check if the server is up
	router.Methods("GET").Path("/healthcheck").HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("ok"))
		})

	// Delegate all of the HTTP routing and serving to the gorilla/mux router.
	// Log all requests using the standard Apache format.
	http.Handle("/", handlers.CombinedLoggingHandler(os.Stderr, router))
}

// appHandler function type
// To reduce the repetition of error handling, we can define our own HTTP appHandler type that includes an error return value
// http://blog.golang.org/error-handling-and-go
type appHandler func(http.ResponseWriter, *http.Request) *common.AppError

// The ServeHTTP method calls the appHandler function and displays the returned error (if any) to the user.
// This is simpler than the original version, but the http package doesn't understand functions that return error.
// To fix this we can implement the http.Handler interface's ServeHTTP method on appHandler
// Notice that the method's receiver, fn, is a function.
// (Go can do that!) The method invokes the function by calling the receiver in the expression fn(w, r).
func (fn appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if e := fn(w, r); e != nil { // e is *common.appError, not os.Error.
		log.Printf("Handler error: status code: %d, message: %s, underlying err: %#v",
			e.Code, e.Message, e.Error)

		http.Error(w, e.Message, e.Code)
	}
}
