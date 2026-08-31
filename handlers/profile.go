package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/yourusername/linkedin-profile-api/cache"
	"github.com/yourusername/linkedin-profile-api/scraper"
)

type ProfileHandler struct {
	client *scraper.Client
	cache  *cache.ProfileCache
}

func NewProfileHandler(client *scraper.Client, cache *cache.ProfileCache) *ProfileHandler {
	return &ProfileHandler{client: client, cache: cache}
}

type errorResponse struct {
	Error string `json:"error"`
}

func (h *ProfileHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	rawURL := r.URL.Query().Get("url")
	if rawURL == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "missing 'url' query parameter"})
		return
	}

	username, err := extractUsername(rawURL)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: err.Error()})
		return
	}

	// Return cached result if available
	if cached, found := h.cache.Get(username); found {
		w.Header().Set("X-Cache", "HIT")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(cached)
		return
	}

	profile, err := h.client.FetchProfile(username)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "not found") {
			w.WriteHeader(http.StatusNotFound)
		} else if strings.Contains(errMsg, "authentication failed") {
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		json.NewEncoder(w).Encode(errorResponse{Error: errMsg})
		return
	}

	h.cache.Set(username, profile)

	w.Header().Set("X-Cache", "MISS")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(profile)
}

// extractUsername pulls the slug from a linkedin.com/in/<slug> URL.
func extractUsername(rawURL string) (string, error) {
	if !strings.HasPrefix(rawURL, "http") {
		rawURL = "https://" + rawURL
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %s", rawURL)
	}

	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(parts) < 2 || parts[0] != "in" || parts[1] == "" {
		return "", fmt.Errorf("URL must be a LinkedIn profile URL like https://linkedin.com/in/username")
	}

	return parts[1], nil
}
