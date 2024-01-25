package provider

import (
	"encoding/json"
	"net/http"
	"strings"
)

func (c *SquaredUpClient) CreateScript(script Script) (*Script, error) {
	rb, err := json.Marshal(script)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.baseURL+"/api/scripts", strings.NewReader(string(rb)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	newScript := Script{}
	err = json.Unmarshal(body, &newScript)
	if err != nil {
		return nil, err
	}

	return &newScript, nil
}

func (c *SquaredUpClient) GetScript(scriptId string) (*Script, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/api/scripts/"+scriptId, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	script := Script{}
	err = json.Unmarshal(body, &script)
	if err != nil {
		return nil, err
	}

	return &script, nil
}

func (c *SquaredUpClient) UpdateScript(scriptId string, script Script) error {
	rb, err := json.Marshal(script)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", c.baseURL+"/api/scripts/"+scriptId, strings.NewReader(string(rb)))
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	if err != nil {
		return err
	}

	return nil
}

func (c *SquaredUpClient) DeleteScript(scriptId string) error {
	req, err := http.NewRequest("DELETE", c.baseURL+"/api/scripts/"+scriptId, nil)
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	if err != nil {
		return err
	}

	return nil
}
