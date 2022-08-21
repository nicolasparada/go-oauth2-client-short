package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/endpoints"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

func run() error {
	var clientID, clientSecret string

	// Get those from registering a new OAuth application in GitHub
	// at https://github.com/settings/applications/new
	// The authorization callback URL should match the configured value down below
	// http://localhost:3000/api/v1/oauth2/github/callback in this case.
	flag.StringVar(&clientID, "client-id", os.Getenv("GITHUB_CLIENT_ID"), "client id")
	flag.StringVar(&clientSecret, "client-secret", os.Getenv("GITHUB_CLIENT_SECRET"), "client secret")
	flag.Parse()

	githubConfig := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     endpoints.GitHub,
		Scopes:       []string{"read:user", "user:email"},
		RedirectURL:  "http://localhost:3000/api/v1/oauth2/github/callback",
	}
	h := &handler{githubConfig: githubConfig}
	http.HandleFunc("/api/v1/oauth2/github/redirect", h.githubRedirect)
	http.HandleFunc("/api/v1/oauth2/github/callback", h.githubCallback)
	return http.ListenAndServe(":3000", nil)
}

type handler struct {
	githubConfig *oauth2.Config
}

func (h *handler) githubRedirect(w http.ResponseWriter, r *http.Request) {
	state := generateState()
	saveState(w, state)
	redirectTo(w, r, h.githubConfig.AuthCodeURL(state))
}

func (h *handler) githubCallback(w http.ResponseWriter, r *http.Request) {
	if !stateOK(r) {
		http.Error(w, "invalid state", http.StatusTeapot)
		return
	}

	ctx := r.Context()
	code := getCode(r)
	token, err := h.githubConfig.Exchange(ctx, code)
	if err != nil {
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
		return
	}

	client := h.githubConfig.Client(ctx, token)
	user, err := fetchUser(ctx, client)
	switch err.(type) {
	case userFetchError:
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
		return
	case userRespError:
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// TODO: save user in session or issue a JWT or whatever you want.
	fmt.Println("user:", user)
}

type userFetchError struct{ error }
type userRespError struct{ error }

func fetchUser(ctx context.Context, client *http.Client) (githubUser, error) {
	var user githubUser

	// Careful, as the /user endpoint doesn't tell if the email is verified
	// and the email could be an empty string too if the user profile is private.
	// Use the /emails endpoints to get the verified email.
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user", nil)
	if err != nil {
		return user, fmt.Errorf("build user request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return user, userFetchError{fmt.Errorf("fetch user: %w", err)}
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return user, userRespError{fmt.Errorf("user respose: %w", err)}
	}

	err = json.NewDecoder(resp.Body).Decode(&user)
	if err != nil {
		return user, fmt.Errorf("json decode user: %w", err)
	}

	return user, nil
}

type githubUser struct {
	// Email field could be empty if the user profile is private.
	Email     string `json:"email"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
}

func generateState() string {
	// TODO: generate a crypto random string.
	return "TODO"
}

func saveState(w http.ResponseWriter, s string) {
	// TODO: save state in a cookie.
}

func redirectTo(w http.ResponseWriter, r *http.Request, u string) {
	http.Redirect(w, r, u, http.StatusFound)
}

func stateOK(r *http.Request) bool {
	// TODO: read saved state from cookie and compare to the one in the request.
	return true
}

func getCode(r *http.Request) string {
	return r.URL.Query().Get("code")
}
