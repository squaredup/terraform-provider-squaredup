package provider

import (
	"encoding/json"
	"net/http"
	"strings"
)

func (c *SquaredUpClient) CreateScope(scope ScopeCreate, workspaceId string) (string, error) {
	rb, err := json.Marshal(scope)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", c.baseURL+"/api/workspaces/"+workspaceId+"/scopes", strings.NewReader(string(rb)))
	if err != nil {
		return "", err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return "", err
	}

	scopeID := string(body)
	scopeID = strings.Trim(scopeID, `"`)

	return scopeID, nil
}

func (c *SquaredUpClient) GetScope(scopeId string, workspaceId string) (*ScopeRead, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/api/workspaces/"+workspaceId+"/scopes/"+scopeId, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	scope := ScopeRead{}
	err = json.Unmarshal(body, &scope)
	if err != nil {
		return nil, err
	}

	return &scope, nil
}

func (c *SquaredUpClient) UpdateScope(scopeId string, scope ScopeCreate, workspaceId string) error {
	rb, err := json.Marshal(scope)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", c.baseURL+"/api/workspaces/"+workspaceId+"/scopes/"+scopeId, strings.NewReader(string(rb)))
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	if err != nil {
		return err
	}

	return nil
}

func (c *SquaredUpClient) DeleteScope(scopeId string, workspaceId string) error {
	req, err := http.NewRequest("DELETE", c.baseURL+"/api/workspaces/"+workspaceId+"/scopes/"+scopeId, nil)
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	if err != nil {
		return err
	}

	return nil
}
