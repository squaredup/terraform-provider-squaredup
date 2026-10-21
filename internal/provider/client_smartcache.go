package provider

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

type SmartCacheDataStream struct {
	ID               string                 `json:"id"`
	PluginConfigID   string                 `json:"pluginConfigId,omitempty"`
	DataSourceConfig map[string]interface{} `json:"dataSourceConfig,omitempty"`
}

type SmartCacheStorage struct {
	SyncMode                string    `json:"syncMode,omitempty"`
	RetainMirrors           *int64    `json:"retainMirrors,omitempty"`
	UniqueIdentifierColumns *[]string `json:"uniqueIdentifierColumns,omitempty"`
	LastModifiedColumn      *string   `json:"lastModifiedColumn,omitempty"`
	RetentionPeriodDays     *int64    `json:"retentionPeriodDays,omitempty"`
}

type SmartCacheCreate struct {
	DisplayName      string               `json:"displayName"`
	PluginConfigID   string               `json:"pluginConfigId"`
	DataStream       SmartCacheDataStream `json:"dataStream"`
	SupportedScope   string               `json:"supportedScope"`
	Scope            interface{}          `json:"scope,omitempty"`
	Timeframe        string               `json:"timeframe,omitempty"`
	FrequencyMinutes *int64               `json:"frequencyMinutes,omitempty"`
	SmartCacheStorage
}

// SmartCachePatch only carries the fields that changed, as a rename or a data stream change
// starts work on the server even when the value sent matches what is stored.
type SmartCachePatch struct {
	DisplayName      *string               `json:"displayName,omitempty"`
	DataStream       *SmartCacheDataStream `json:"dataStream,omitempty"`
	Scope            interface{}           `json:"scope,omitempty"`
	Timeframe        *string               `json:"timeframe,omitempty"`
	FrequencyMinutes *int64                `json:"frequencyMinutes,omitempty"`
	SmartCacheStorage
}

type SmartCacheStep struct {
	DataStream        SmartCacheDataStream `json:"dataStream"`
	Timeframe         interface{}          `json:"timeframe"`
	SupportedScope    string               `json:"supportedScope"`
	SmartCacheMapping struct {
		TableName string `json:"tableName"`
		SmartCacheStorage
	} `json:"smartCacheMapping"`
}

type SmartCacheRead struct {
	ID               string           `json:"id"`
	Name             string           `json:"name"`
	DisplayName      string           `json:"displayName"`
	FrequencyMinutes int64            `json:"frequencyMinutes"`
	Enabled          bool             `json:"enabled"`
	Steps            []SmartCacheStep `json:"steps"`
}

func (c *SquaredUpClient) CreateSmartCache(smartCache SmartCacheCreate) (string, error) {
	rb, err := json.Marshal(smartCache)
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
		SmartCacheDefinitionID string `json:"smartCacheDefinitionId"`
	}{}
	err = json.Unmarshal(body, &created)
	if err != nil {
		return "", err
	}

	return created.SmartCacheDefinitionID, nil
}

// GetSmartCache returns nil without an error when the SmartCache does not exist.
func (c *SquaredUpClient) GetSmartCache(smartCacheId string) (*SmartCacheRead, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/api/smartcache/"+smartCacheId, nil)
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

	smartCache := SmartCacheRead{}
	err = json.Unmarshal(body, &smartCache)
	if err != nil {
		return nil, err
	}

	return &smartCache, nil
}

func (c *SquaredUpClient) PatchSmartCache(smartCacheId string, patch SmartCachePatch) error {
	rb, err := json.Marshal(patch)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PATCH", c.baseURL+"/api/smartcache/"+smartCacheId, strings.NewReader(string(rb)))
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	return err
}

func (c *SquaredUpClient) SetSmartCacheEnabled(smartCacheId string, enabled bool) error {
	rb, err := json.Marshal(map[string]bool{"enabled": enabled})
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", c.baseURL+"/api/smartcache/"+smartCacheId, strings.NewReader(string(rb)))
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	return err
}

func (c *SquaredUpClient) DeleteSmartCache(smartCacheId string) error {
	req, err := http.NewRequest("DELETE", c.baseURL+"/api/smartcache/"+smartCacheId, nil)
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	return err
}
