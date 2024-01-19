package provider

import (
	"encoding/json"
	"net/http"
	"strings"
)

func (c *SquaredUpClient) CreateSharedDashboard(dashboardShare DashboardShare) (*DashboardShare, error) {
	rb, err := json.Marshal(dashboardShare)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.baseURL+"/api/openaccess/shares", strings.NewReader(string(rb)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	sharedDashboard := DashboardShare{}
	err = json.Unmarshal(body, &sharedDashboard)
	if err != nil {
		return nil, err
	}

	return &sharedDashboard, nil
}

func (c *SquaredUpClient) GetSharedDashboard(sharedDashboardId string) (*DashboardShare, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/api/openaccess/shares/"+sharedDashboardId, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	sharedDashboard := DashboardShare{}
	err = json.Unmarshal(body, &sharedDashboard)
	if err != nil {
		return nil, err
	}

	return &sharedDashboard, nil
}

func (c *SquaredUpClient) UpdateSharedDashboard(sharedDashboardId string, dashboardShare DashboardShare) error {
	rb, err := json.Marshal(dashboardShare)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", c.baseURL+"/api/openaccess/shares/"+sharedDashboardId, strings.NewReader(string(rb)))
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	if err != nil {
		return err
	}

	return nil
}

func (c *SquaredUpClient) DeleteSharedDashboard(sharedDashboardId string) error {
	req, err := http.NewRequest("DELETE", c.baseURL+"/api/openaccess/shares/"+sharedDashboardId, nil)
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	if err != nil {
		return err
	}

	return nil
}
