package redis

import "time"

const (
	flagCachePrefix        = "flag:"
	apiKeyCachePrefix      = "apikey:"
	tenantCachePrefix      = "tenant:"
	applicationCachePrefix = "app:"
	environmentCachePrefix = "env:"
	userCachePrefix        = "user:"
	defaultTTL             = 5 * time.Minute
	shortTTL               = 1 * time.Minute
	longTTL                = 30 * time.Minute
)
