package provider

import (
	"encoding/json"
	"net/http"
	"strings"
)

func (c *SquaredUpClient) CreateWorkspace(workspacePayload map[string]interface{}) (string, error) {
	rb, err := json.Marshal(workspacePayload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", c.baseURL+"/api/workspaces", strings.NewReader(string(rb)))
	if err != nil {
		return "", err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return "", err
	}

	workspaceID := string(body)

	return workspaceID, nil
}

func (c *SquaredUpClient) GetWorkspace(workspaceId string) (*WorkspaceRead, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/api/workspaces/"+workspaceId, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	workspace := WorkspaceRead{}
	err = json.Unmarshal(body, &workspace)
	if err != nil {
		return nil, err
	}

	return &workspace, nil
}

func (c *SquaredUpClient) UpdateWorkspace(workspaceId string, workspacePayload map[string]interface{}) error {
	rb, err := json.Marshal(workspacePayload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", c.baseURL+"/api/workspaces/"+workspaceId, strings.NewReader(string(rb)))
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	if err != nil {
		return err
	}

	return nil
}

func (c *SquaredUpClient) DeleteWorkspace(workspaceId string) error {
	req, err := http.NewRequest("DELETE", c.baseURL+"/api/workspaces/"+workspaceId, nil)
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	if err != nil {
		return err
	}

	return nil
}
