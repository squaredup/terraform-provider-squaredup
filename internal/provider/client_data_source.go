package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func (c *SquaredUpClient) GetLatestDataSources(filterDisplayName string) ([]LatestDataSource, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/api/plugins/latest", nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	plugins := []LatestDataSource{}
	err = json.Unmarshal(body, &plugins)
	if err != nil {
		return nil, err
	}

	if filterDisplayName != "" {
		filteredPlugins := []LatestDataSource{}
		for _, plugin := range plugins {
			if plugin.DisplayName == filterDisplayName {
				filteredPlugins = append(filteredPlugins, plugin)
			}
		}

		if len(filteredPlugins) == 0 {
			return nil, fmt.Errorf("No plugins found with display name: %s", filterDisplayName)
		}

		return filteredPlugins, nil
	}

	return plugins, nil
}

func (c *SquaredUpClient) GenerateDataSourcePayload(displayName string, name string, pluginConfig map[string]interface{}, secureJsonData map[string]interface{}, agentGroupId string) (map[string]interface{}, error) {
	plugins, err := c.GetLatestDataSources(name)
	if err != nil {
		return nil, err
	}

	DataSourcePayload := map[string]interface{}{
		"displayName": displayName,
		"config": map[string]interface{}{
			"pluginId":   plugins[0].PluginID,
			"lambdaName": plugins[0].LambdaName,
			"version":    plugins[0].Version,
		},
		"plugin": map[string]interface{}{
			"pluginId":           plugins[0].PluginID,
			"name":               plugins[0].DisplayName,
			"lambdaName":         plugins[0].LambdaName,
			"displayName":        plugins[0].DisplayName,
			"version":            plugins[0].Version,
			"onPrem":             plugins[0].OnPrem,
			"importNotSupported": false,
		},
		"agentGroupId": agentGroupId,
	}

	for key, value := range pluginConfig {
		config, ok := DataSourcePayload["config"].(map[string]interface{})
		if !ok {
			config = make(map[string]interface{})
			DataSourcePayload["config"] = config
		}
		config[key] = value
	}

	for key, value := range secureJsonData {
		config, ok := DataSourcePayload["config"].(map[string]interface{})
		if !ok {
			config = make(map[string]interface{})
			DataSourcePayload["config"] = config
		}
		config[key] = value
	}

	return DataSourcePayload, nil
}

func (c *SquaredUpClient) AddDataSource(displayName string, name string, pluginConfig map[string]interface{}, secureJsonData map[string]interface{}, agentGroupId string) (*DataSource, error) {
	DataSourcePayload, err := c.GenerateDataSourcePayload(displayName, name, pluginConfig, secureJsonData, agentGroupId)
	if err != nil {
		return nil, err
	}

	rb, err := json.Marshal(DataSourcePayload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.baseURL+"/api/source/configs", strings.NewReader(string(rb)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	newDataSource := DataSource{}
	err = json.Unmarshal(body, &newDataSource)
	if err != nil {
		return nil, err
	}

	return &newDataSource, nil
}

func (c *SquaredUpClient) GetDataSource(dataSourceId string) (*DataSource, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/api/source/configs/"+dataSourceId, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	dataSource := DataSource{}
	err = json.Unmarshal(body, &dataSource)
	if err != nil {
		return nil, err
	}

	return &dataSource, nil
}

func (c *SquaredUpClient) UpdateDataSource(dataSourceId string, displayName string, name string, pluginConfig map[string]interface{}, secureJsonData map[string]interface{}, agentGroupId string) error {
	DataSourcePayload, err := c.GenerateDataSourcePayload(displayName, name, pluginConfig, secureJsonData, agentGroupId)
	if err != nil {
		return err
	}

	rb, err := json.Marshal(DataSourcePayload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", c.baseURL+"/api/source/configs/"+dataSourceId, strings.NewReader(string(rb)))
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	if err != nil {
		return err
	}

	return nil
}

func (c *SquaredUpClient) DeleteDataSource(dataSourceId string) error {
	req, err := http.NewRequest("DELETE", c.baseURL+"/api/source/configs/"+dataSourceId, nil)
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	if err != nil {
		return err
	}

	return nil
}
