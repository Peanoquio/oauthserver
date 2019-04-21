# oauthserver
The oauth server which will manage authentication to various platforms such as Google, Facebook, etc.
This will also support session creation, validation and termination.

---

## Tutorial

Before starting, run this command to ensure that the Go module dependencies are updated
```bash
go mod tidy
```

Create the required environment variables and place them in a file called `env`
For Google Oauth2, you need to go to the Google developer site to create the client ID and secret for your app
```bash
export GOOGLE_OAUTH2_CLIENTID=<some value>
export GOOGLE_OAUTH2_CLIENTSECRET=<some value>
export COOKIE_STORE_SECRET=<some value>
```

Load the environment variables
```bash
source ./env
```

Start the server
```bash
go start server.go
```

Open your browser and try navigating to this page.
You will see that you are not yet authorized to access this page.
```
http://localhost:8080/test
```
Navigate to this page and login. You'll be prompted to enter your Google credentials.
```
http://localhost:8080/oauthpage
```
After being authenticated by Google, you should be able to access the test page.
You can also play with the behaviour by logging out then check to see if you can access the test page.

---

## TODO
This is built for web page/app authentication--needs to be tested on mobile apps.
Moreover, for security--PKCE needs to be implemented which is much needed by mobile or native apps.
