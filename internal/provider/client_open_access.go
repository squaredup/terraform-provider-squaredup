package provider

import (
	"encoding/json"
	"net/http"
	"strings"
)

func (c *SquaredUpClient) CreateOpenAccess(openAccess OpenAccess) (*OpenAccess, error) {
	rb, err := json.Marshal(openAccess)
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

	newOpenAccess := OpenAccess{}
	err = json.Unmarshal(body, &newOpenAccess)
	if err != nil {
		return nil, err
	}

	return &newOpenAccess, nil
}

func (c *SquaredUpClient) GetOpenAccess(openAcessId string) (*OpenAccess, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/api/openaccess/shares/"+openAcessId, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	openAccess := OpenAccess{}
	err = json.Unmarshal(body, &openAccess)
	if err != nil {
		return nil, err
	}

	return &openAccess, nil
}

func (c *SquaredUpClient) UpdateOpenAccess(openAcessId string, openAccess OpenAccess) error {
	rb, err := json.Marshal(openAccess)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", c.baseURL+"/api/openaccess/shares/"+openAcessId, strings.NewReader(string(rb)))
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	if err != nil {
		return err
	}

	return nil
}

func (c *SquaredUpClient) DeleteOpenAccess(openAcessId string) error {
	req, err := http.NewRequest("DELETE", c.baseURL+"/api/openaccess/shares/"+openAcessId, nil)
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	if err != nil {
		return err
	}

	return nil
}
