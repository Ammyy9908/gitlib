package github

import (
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"net/http"
)

var (
	oauthConf = &oauth2.Config{
		ClientID:     "804029bf157ab09a26f9",                     // Fill in your Client ID
		ClientSecret: "f081cba4a12ef4fa5a38668f58b2d4bd2706da5a", // Fill in your Client Secret
		Scopes:       []string{"repo", "user"},
		Endpoint:     github.Endpoint,
	}
	// Random string for state to prevent CSRF
	oauthStateString = "random" // You should generate a random string for production
)

func HandleGitHubLogin(w http.ResponseWriter, r *http.Request) {
	url := oauthConf.AuthCodeURL(oauthStateString, oauth2.AccessTypeOnline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}
