package view

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"path/filepath"

	"github.com/Peanoquio/oauthserver/common"
)

const (
	// baseHTML is the filename of the base HMTL file which will be appended with other HTML files
	baseHTML = "base.html"
	// templateDir is the directory where template HTML files are stored
	templateDir = "view/templates"
	// templateTag is the name of the tag in the parent HTML where the child HMTL will be appended to
	templateTag = "body"
)

// ParseTemplate applies a given file to the body of the base template.
func ParseTemplate(filename string) *AppTemplate {
	tmpl := template.Must(template.ParseFiles(templateDir + "/" + baseHTML))

	// Put the named file into a template called "body"
	path := filepath.Join(templateDir, filename)
	b, err := ioutil.ReadFile(path)
	if err != nil {
		panic(fmt.Errorf("could not read template: %v", err))
	}
	template.Must(tmpl.New(templateTag).Parse(string(b)))

	return &AppTemplate{tmpl.Lookup(baseHTML)}
}

// AppTemplate is a user login-aware wrapper for a html/template.
type AppTemplate struct {
	template *template.Template
}

// Execute writes the template using the provided data, adding login and user information to the base template.
func (appTmpl *AppTemplate) Execute(w http.ResponseWriter, r *http.Request, isAuthEnabled bool, profile *common.Profile, data interface{}) *common.AppError {
	tmplData := struct {
		Data        interface{}
		AuthEnabled bool
		Profile     *common.Profile
		LoginURL    string
		LogoutURL   string
	}{
		Data:        data,
		AuthEnabled: isAuthEnabled,
		LoginURL:    "/login?redirect=" + r.URL.RequestURI(),
		LogoutURL:   "/logout?redirect=" + r.URL.RequestURI(),
	}

	if tmplData.AuthEnabled {
		// Ignore any errors.
		tmplData.Profile = profile
	}

	if err := appTmpl.template.Execute(w, tmplData); err != nil {
		return common.AppErrorf(err, "could not write template: %v", err)
	}
	return nil
}
