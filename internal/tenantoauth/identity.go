package tenantoauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"golang.org/x/oauth2"
)

func (s *Service) fetchIdentity(ctx context.Context, provider string, token *oauth2.Token) (Identity, error) {
	switch strings.ToLower(provider) {
	case "google":
		return fetchGoogleIdentity(ctx, token)
	case "github":
		return fetchGitHubIdentity(ctx, token)
	default:
		return Identity{}, fmt.Errorf("unsupported provider")
	}
}

func fetchGoogleIdentity(ctx context.Context, token *oauth2.Token) (Identity, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return Identity{}, err
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return Identity{}, err
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		return Identity{}, fmt.Errorf("google userinfo status %d", res.StatusCode)
	}
	var payload struct {
		ID      string `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Verified bool  `json:"verified_email"`
	}
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return Identity{}, err
	}
	if payload.ID == "" || payload.Email == "" {
		return Identity{}, fmt.Errorf("google profile incomplete")
	}
	if !payload.Verified {
		return Identity{}, fmt.Errorf("google email is not verified")
	}
	return Identity{
		Provider:       "google",
		ProviderUserID: payload.ID,
		Email:          strings.ToLower(strings.TrimSpace(payload.Email)),
		DisplayName:    strings.TrimSpace(payload.Name),
	}, nil
}

func fetchGitHubIdentity(ctx context.Context, token *oauth2.Token) (Identity, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user", nil)
	if err != nil {
		return Identity{}, err
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	req.Header.Set("Accept", "application/vnd.github+json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return Identity{}, err
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		return Identity{}, fmt.Errorf("github user status %d", res.StatusCode)
	}
	var user struct {
		ID    int64  `json:"id"`
		Login string `json:"login"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	if err := json.NewDecoder(res.Body).Decode(&user); err != nil {
		return Identity{}, err
	}
	email := strings.ToLower(strings.TrimSpace(user.Email))
	if email == "" {
		email, err = fetchGitHubPrimaryEmail(ctx, token)
		if err != nil {
			return Identity{}, err
		}
	}
	display := strings.TrimSpace(user.Name)
	if display == "" {
		display = user.Login
	}
	return Identity{
		Provider:       "github",
		ProviderUserID: fmt.Sprintf("%d", user.ID),
		Email:          email,
		DisplayName:    display,
	}, nil
}

func fetchGitHubPrimaryEmail(ctx context.Context, token *oauth2.Token) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user/emails", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	req.Header.Set("Accept", "application/vnd.github+json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		return "", fmt.Errorf("github emails status %d", res.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return "", err
	}
	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}
	if err := json.Unmarshal(body, &emails); err != nil {
		return "", err
	}
	for _, item := range emails {
		if item.Primary && item.Verified && strings.TrimSpace(item.Email) != "" {
			return strings.ToLower(strings.TrimSpace(item.Email)), nil
		}
	}
	for _, item := range emails {
		if item.Verified && strings.TrimSpace(item.Email) != "" {
			return strings.ToLower(strings.TrimSpace(item.Email)), nil
		}
	}
	return "", fmt.Errorf("github email not available")
}