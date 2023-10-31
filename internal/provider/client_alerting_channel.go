package provider

import (
	"encoding/json"
	"net/http"
	"strings"
)

func (c *SquaredUpClient) CreateAlertingChannel(alertChannel AlertingChannel) (*AlertingChannel, error) {
	rb, err := json.Marshal(alertChannel)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.baseURL+"/api/alerting/channels", strings.NewReader(string(rb)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	newAlertChannel := AlertingChannel{}
	err = json.Unmarshal(body, &newAlertChannel)
	if err != nil {
		return nil, err
	}

	return &newAlertChannel, nil
}

func (c *SquaredUpClient) GetAlertingChannel(alertChannelId string) (*AlertingChannel, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/api/alerting/channels/"+alertChannelId, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	alertChannel := AlertingChannel{}
	err = json.Unmarshal(body, &alertChannel)
	if err != nil {
		return nil, err
	}

	return &alertChannel, nil
}

func (c *SquaredUpClient) UpdateAlertingChannel(alertChannelId string, alertChannel AlertingChannel) error {
	rb, err := json.Marshal(alertChannel)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", c.baseURL+"/api/alerting/channels/"+alertChannelId, strings.NewReader(string(rb)))
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	if err != nil {
		return err
	}

	return nil
}

func (c *SquaredUpClient) DeleteAlertingChannel(alertChannelId string) error {
	req, err := http.NewRequest("DELETE", c.baseURL+"/api/alerting/channels/"+alertChannelId, nil)
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	if err != nil {
		return err
	}

	return nil
}
