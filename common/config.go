package common

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/go-yaml/yaml"
)

// GoogleOauth2Callback is the callback registered in Google
// After the user authenticates through Goolge login, Google will redirect to this callback URL
const GoogleOauth2Callback = "http://localhost:8080/oauth2callback"

// EnvVars will be based on the environment variables
type EnvVars struct {
	GoogleOauth2ClientID     string `yaml:"GOOGLE_OAUTH2_CLIENTID"`
	GoogleOauth2ClientSecret string `yaml:"GOOGLE_OAUTH2_CLIENTSECRET"`
	GoogleOauth2Callback     string `yaml:"GOOGLE_OAUTH2_CALLBACK"`
	CookieStoreSecret        string `yaml:"COOKIE_STORE_SECRET"`
}

// LoadEnvironmentVariables loads the environment variables based on the yaml file and stores it in common.EnvVars
func LoadEnvironmentVariables() *EnvVars {
	confContent, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		panic(err)
	}
	// expand environment variables
	confContent = []byte(os.ExpandEnv(string(confContent)))
	envConfig := &EnvVars{}
	if err := yaml.Unmarshal(confContent, envConfig); err != nil {
		panic(err)
	}
	if envConfig.GoogleOauth2Callback == "" {
		envConfig.GoogleOauth2Callback = GoogleOauth2Callback
	}
	fmt.Printf("environment variables loaded: %+v\n", envConfig)
	return envConfig
}
