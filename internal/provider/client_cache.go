package provider

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

type CacheDataStream struct {
	ID               string                 `json:"id"`
	PluginConfigID   string                 `json:"pluginConfigId,omitempty"`
	DataSourceConfig map[string]interface{} `json:"dataSourceConfig,omitempty"`
}

type CacheCreate struct {
	DisplayName      string          `json:"displayName"`
	PluginConfigID   string          `json:"pluginConfigId"`
	DataStream       CacheDataStream `json:"dataStream"`
	SupportedScope   string          `json:"supportedScope"`
	Scope            interface{}     `json:"scope,omitempty"`
	Timeframe        string          `json:"timeframe,omitempty"`
	FrequencyMinutes *int64          `json:"frequencyMinutes,omitempty"`
	// Caches only support full sync
	SyncMode string `json:"syncMode"`
}

// CachePatch only carries the fields that changed, as a rename or a data stream change
// starts work on the server even when the value sent matches what is stored.
type CachePatch struct {
	DisplayName      *string          `json:"displayName,omitempty"`
	DataStream       *CacheDataStream `json:"dataStream,omitempty"`
	Scope            interface{}      `json:"scope,omitempty"`
	Timeframe        *string          `json:"timeframe,omitempty"`
	FrequencyMinutes *int64           `json:"frequencyMinutes,omitempty"`
}

type CacheStep struct {
	DataStream     CacheDataStream `json:"dataStream"`
	Timeframe      interface{}     `json:"timeframe"`
	SupportedScope string          `json:"supportedScope"`
	Scope          interface{}     `json:"scope"`
	CacheMapping   struct {
		TableName string `json:"tableName"`
		SyncMode  string `json:"syncMode"`
	} `json:"smartCacheMapping"`
}

type CacheRead struct {
	ID               string      `json:"id"`
	Name             string      `json:"name"`
	DisplayName      string      `json:"displayName"`
	FrequencyMinutes int64       `json:"frequencyMinutes"`
	Enabled          bool        `json:"enabled"`
	Steps            []CacheStep `json:"steps"`
}

func (c *SquaredUpClient) CreateCache(cache CacheCreate) (string, error) {
	rb, err := json.Marshal(cache)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", c.baseURL+"/api/smartcache", strings.NewReader(string(rb)))
	if err != nil {
		return "", err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return "", err
	}

	created := struct {
		CacheDefinitionID string `json:"smartCacheDefinitionId"`
	}{}
	err = json.Unmarshal(body, &created)
	if err != nil {
		return "", err
	}

	return created.CacheDefinitionID, nil
}

// GetCache returns nil without an error when the cache does not exist.
func (c *SquaredUpClient) GetCache(cacheId string) (*CacheRead, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/api/smartcache/"+cacheId, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	var apiErr *APIError
	if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	cache := CacheRead{}
	err = json.Unmarshal(body, &cache)
	if err != nil {
		return nil, err
	}

	return &cache, nil
}

func (c *SquaredUpClient) PatchCache(cacheId string, patch CachePatch) error {
	rb, err := json.Marshal(patch)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PATCH", c.baseURL+"/api/smartcache/"+cacheId, strings.NewReader(string(rb)))
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	return err
}

func (c *SquaredUpClient) SetCacheEnabled(cacheId string, enabled bool) error {
	rb, err := json.Marshal(map[string]bool{"enabled": enabled})
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", c.baseURL+"/api/smartcache/"+cacheId, strings.NewReader(string(rb)))
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	return err
}

func (c *SquaredUpClient) DeleteCache(cacheId string) error {
	req, err := http.NewRequest("DELETE", c.baseURL+"/api/smartcache/"+cacheId, nil)
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	return err
}

type dataStreamPlugin struct {
	PluginID       string `json:"pluginId"`
	DataSourceName string `json:"dataSourceName"`
}

type pluginDataSources struct {
	DataSources []struct {
		Name           string `json:"name"`
		SupportedScope string `json:"supportedScope"`
	} `json:"dataSources"`
}

// GetDataStreamSupportedScope returns how a data stream accepts objects (single, list or none), as
// declared by the plugin function behind it. The SquaredUp UI reads it the same way.
func (c *SquaredUpClient) GetDataStreamSupportedScope(dataStreamId string) (string, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/api/datastreams/"+dataStreamId, nil)
	if err != nil {
		return "", err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return "", err
	}

	dataStream := dataStreamPlugin{}
	err = json.Unmarshal(body, &dataStream)
	if err != nil {
		return "", err
	}

	req, err = http.NewRequest("GET", c.baseURL+"/api/plugins/"+dataStream.PluginID, nil)
	if err != nil {
		return "", err
	}

	body, err = c.doRequest(req)
	if err != nil {
		return "", err
	}

	plugin := pluginDataSources{}
	err = json.Unmarshal(body, &plugin)
	if err != nil {
		return "", err
	}

	return findSupportedScope(plugin, dataStream)
}

func findSupportedScope(plugin pluginDataSources, dataStream dataStreamPlugin) (string, error) {
	for _, dataSource := range plugin.DataSources {
		if dataSource.Name == dataStream.DataSourceName && dataSource.SupportedScope != "" {
			return dataSource.SupportedScope, nil
		}
	}

	return "", fmt.Errorf("plugin %s declares no supported scope for %q", dataStream.PluginID, dataStream.DataSourceName)
}
