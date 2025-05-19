package provider

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func (c *SquaredUpClient) GetDashboardImage(spaceId, dashId, tileId string) (*DashboardImage, error) {
	currentTimeInMillis := time.Now().UnixMilli()
	currentTimeStr := strconv.FormatInt(currentTimeInMillis, 10)
	url := c.baseURL + "/api/workspaces/" + spaceId + "/dashboards/" + dashId + "/images/" + tileId + "?uploaded=" + currentTimeStr
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	dashboardImage := DashboardImage{}
	err = json.Unmarshal(body, &dashboardImage)
	if err != nil {
		return nil, err
	}

	return &dashboardImage, nil
}

func (c *SquaredUpClient) UploadDashboardImage(spaceId, dashId, tileId string, dashboardImage *DashboardImage) error {
	url := c.baseURL + "/api/workspaces/" + spaceId + "/dashboards/" + dashId + "/images/" + tileId
	rb, err := json.Marshal(dashboardImage)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", url, strings.NewReader(string(rb)))
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	if err != nil {
		return err
	}

	return nil
}

func (c *SquaredUpClient) DeleteDashboardImage(spaceId, dashId, tileId string) error {
	url := c.baseURL + "/api/workspaces/" + spaceId + "/dashboards/" + dashId + "/images/" + tileId
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	if err != nil {
		return err
	}

	return nil
}
