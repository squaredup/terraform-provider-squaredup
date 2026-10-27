package provider

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

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

// GetDataStreamSupportedScope returns how a data stream accepts objects (single, list or none), as
// declared by the plugin function behind it. The SquaredUp UI reads it the same way.
func (c *SquaredUpClient) GetDataStreamSupportedScope(dataStreamId string) (string, error) {
	dataStream, err := c.getDataStreamPlugin(dataStreamId)
	if err != nil {
		return "", err
	}

	plugin, err := c.getPluginDataSources(dataStream.PluginID)
	if err != nil {
		return "", err
	}

	return findSupportedScope(plugin, dataStream)
}

// The lookups are cached for the life of the provider, a single plan or apply, so caches sharing a
// data stream or plugin don't each fetch it.
func (c *SquaredUpClient) getDataStreamPlugin(dataStreamId string) (dataStreamPlugin, error) {
	if cached, ok := c.dataStreamPlugins.Load(dataStreamId); ok {
		if value, ok := cached.(dataStreamPlugin); ok {
			return value, nil
		}
	}

	req, err := http.NewRequest("GET", c.baseURL+"/api/datastreams/"+dataStreamId, nil)
	if err != nil {
		return dataStreamPlugin{}, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return dataStreamPlugin{}, err
	}

	dataStream := dataStreamPlugin{}
	err = json.Unmarshal(body, &dataStream)
	if err != nil {
		return dataStreamPlugin{}, err
	}

	c.dataStreamPlugins.Store(dataStreamId, dataStream)
	return dataStream, nil
}

func (c *SquaredUpClient) getPluginDataSources(pluginId string) (pluginDataSources, error) {
	if cached, ok := c.pluginDataSources.Load(pluginId); ok {
		if value, ok := cached.(pluginDataSources); ok {
			return value, nil
		}
	}

	req, err := http.NewRequest("GET", c.baseURL+"/api/plugins/"+pluginId, nil)
	if err != nil {
		return pluginDataSources{}, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return pluginDataSources{}, err
	}

	plugin := pluginDataSources{}
	err = json.Unmarshal(body, &plugin)
	if err != nil {
		return pluginDataSources{}, err
	}

	c.pluginDataSources.Store(pluginId, plugin)
	return plugin, nil
}

func findSupportedScope(plugin pluginDataSources, dataStream dataStreamPlugin) (string, error) {
	for _, dataSource := range plugin.DataSources {
		if dataSource.Name == dataStream.DataSourceName && dataSource.SupportedScope != "" {
			return dataSource.SupportedScope, nil
		}
	}

	return "", fmt.Errorf("plugin %s declares no supported scope for %q", dataStream.PluginID, dataStream.DataSourceName)
}
