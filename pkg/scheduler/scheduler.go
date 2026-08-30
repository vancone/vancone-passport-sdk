package scheduler

import (
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/vancone/vancone-passport-sdk-go/pkg/config"
	"github.com/vancone/vancone-passport-sdk-go/pkg/constant"
	"github.com/vancone/vancone-passport-sdk-go/pkg/model"
	"github.com/vancone/vancone-passport-sdk-go/pkg/util"
	"github.com/vancone/vancone-web-common-go/pkg/logger"
	"github.com/vancone/vancone-web-common-go/pkg/server/response"
)

type compiledApi struct {
	api     model.Api
	pattern *util.PathPattern
}

var (
	startOnce       sync.Once
	cacheMux        sync.RWMutex
	patternParser   = util.NewPathPatternParser()
	apiCache        []model.Api
	apiPatterns     []compiledApi
	permissionCache []model.Permission
)

// Start launches the background goroutine that periodically syncs the API and
// permission caches from the Passport platform.
func Start() {
	startOnce.Do(func() {
		go func() {
			SyncApis()
			SyncPermissions()
			period := config.LocalConfig.Cache.SyncPeriodSeconds
			if period <= 0 {
				period = 60
			}
			ticker := time.NewTicker(time.Duration(period) * time.Second)
			defer ticker.Stop()
			for range ticker.C {
				SyncApis()
				SyncPermissions()
			}
		}()
	})
}

func SyncApis() {
	respBytes := util.RequestWithAuth(config.LocalConfig.BaseUrl+constant.SyncApiUrl, "GET", "")
	resp := response.Response[[]model.Api]{}
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		logger.Error("Failed to sync API list", err)
		return
	}
	if resp.Code != 0 {
		logger.Error("Failed to sync API list", errors.New(resp.Message))
		return
	}
	patterns := make([]compiledApi, 0, len(resp.Data))
	for _, api := range resp.Data {
		pattern, err := patternParser.Parse(api.Path)
		if err != nil {
			logger.Error("Failed to parse API path pattern", api.Path, err)
			continue
		}
		patterns = append(patterns, compiledApi{api: api, pattern: pattern})
	}
	cacheMux.Lock()
	apiCache = resp.Data
	apiPatterns = patterns
	cacheMux.Unlock()
}

func SyncPermissions() {
	respBytes := util.RequestWithAuth(config.LocalConfig.BaseUrl+constant.SyncPermissionUrl, "GET", "")
	resp := response.Response[[]model.Permission]{}
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		logger.Error("Failed to sync permission list", err)
		return
	}
	if resp.Code != 0 {
		logger.Error("Failed to sync permission list", errors.New(resp.Message))
		return
	}
	cacheMux.Lock()
	permissionCache = resp.Data
	cacheMux.Unlock()
}

func GetApiCache() []model.Api {
	cacheMux.RLock()
	defer cacheMux.RUnlock()
	return apiCache
}

func GetPermissionCache() []model.Permission {
	cacheMux.RLock()
	defer cacheMux.RUnlock()
	return permissionCache
}

// MatchApi finds the first registered API whose path pattern and method match
// the request, returning nil when no API matches.
func MatchApi(path string, method string) *model.Api {
	cacheMux.RLock()
	defer cacheMux.RUnlock()
	for _, compiled := range apiPatterns {
		if compiled.pattern.RegexPattern.MatchString(path) && compiled.api.Method == method {
			api := compiled.api
			return &api
		}
	}
	return nil
}
