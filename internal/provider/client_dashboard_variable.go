package provider

import (
	"encoding/json"
	"net/http"
	"strings"
)

func (c *SquaredUpClient) CreateDashboardVariable(variable DashboardVariable, workspaceId string) (*DashboardVariableRead, error) {
	rb, err := json.Marshal(variable)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.baseURL+"/api/workspaces/"+workspaceId+"/variables", strings.NewReader(string(rb)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	variableRead := DashboardVariableRead{}
	err = json.Unmarshal(body, &variableRead)
	if err != nil {
		return nil, err
	}

	return &variableRead, nil
}

func (c *SquaredUpClient) GetDashboardVariable(variableId string) (*DashboardVariableRead, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/api/variables/"+variableId, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	variable := DashboardVariableRead{}
	err = json.Unmarshal(body, &variable)
	if err != nil {
		return nil, err
	}

	return &variable, nil
}

func (c *SquaredUpClient) UpdateDashboardVariable(variableId string, variable DashboardVariable) (*DashboardVariableRead, error) {
	rb, err := json.Marshal(variable)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", c.baseURL+"/api/variables/"+variableId, strings.NewReader(string(rb)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	variableRead := DashboardVariableRead{}
	err = json.Unmarshal(body, &variableRead)
	if err != nil {
		return nil, err
	}

	return &variableRead, nil
}

func (c *SquaredUpClient) DeleteDashboardVariable(variableId string) error {
	req, err := http.NewRequest("DELETE", c.baseURL+"/api/variables/"+variableId, nil)
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	if err != nil {
		return err
	}

	return nil
}
