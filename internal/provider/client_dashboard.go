package provider

import (
	"encoding/json"
	"net/http"
	"strings"
)

func (c *SquaredUpClient) CreateDashboard(displayName string, workspaceId string, timeframe string, dashboardContent string) (*Dashboard, error) {

	DashboardPayload := map[string]interface{}{
		"displayName": displayName,
		"workspaceId": workspaceId,
		"timeframe":   timeframe,
	}

	rb, err := json.Marshal(DashboardPayload)
	if err != nil {
		return nil, err
	}

	rb = []byte(strings.Replace(string(rb), "}", ",\"content\":"+dashboardContent+"}", 1))

	req, err := http.NewRequest("POST", c.baseURL+"/api/dashboards", strings.NewReader(string(rb)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	newDashboard := Dashboard{}
	err = json.Unmarshal(body, &newDashboard)
	if err != nil {
		return nil, err
	}

	return &newDashboard, nil
}

func (c *SquaredUpClient) GetDashboard(dashboardId string) (*Dashboard, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/api/dashboards/"+dashboardId, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	newDashboard := Dashboard{}
	err = json.Unmarshal(body, &newDashboard)
	if err != nil {
		return nil, err
	}

	return &newDashboard, nil
}

func (c *SquaredUpClient) UpdateDashboard(dashboardId string, displayName string, timeframe string, dashboardContent string) (*Dashboard, error) {
	DashboardPayload := map[string]interface{}{
		"displayName": displayName,
		"timeframe":   timeframe,
	}

	rb, err := json.Marshal(DashboardPayload)
	if err != nil {
		return nil, err
	}

	rb = []byte(strings.Replace(string(rb), "}", ",\"content\":"+dashboardContent+"}", 1))

	req, err := http.NewRequest("PUT", c.baseURL+"/api/dashboards/"+dashboardId, strings.NewReader(string(rb)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	newDashboard := Dashboard{}
	err = json.Unmarshal(body, &newDashboard)
	if err != nil {
		return nil, err
	}

	return &newDashboard, nil
}

func (c *SquaredUpClient) DeleteDashboard(dashboardId string) error {
	req, err := http.NewRequest("DELETE", c.baseURL+"/api/dashboards/"+dashboardId, nil)
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	if err != nil {
		return err
	}

	return nil
}
