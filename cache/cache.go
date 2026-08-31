package cache

import (
	"time"

	gocache "github.com/patrickmn/go-cache"
	"github.com/yourusername/linkedin-profile-api/models"
)

const (
	defaultTTL     = 30 * time.Minute
	cleanupInterval = 10 * time.Minute
)

type ProfileCache struct {
	store *gocache.Cache
}

func NewProfileCache() *ProfileCache {
	return &ProfileCache{
		store: gocache.New(defaultTTL, cleanupInterval),
	}
}

func (c *ProfileCache) Get(username string) (*models.Profile, bool) {
	val, found := c.store.Get(username)
	if !found {
		return nil, false
	}
	profile, ok := val.(*models.Profile)
	return profile, ok
}

func (c *ProfileCache) Set(username string, profile *models.Profile) {
	c.store.Set(username, profile, defaultTTL)
}
