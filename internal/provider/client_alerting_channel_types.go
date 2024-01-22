package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (c *SquaredUpClient) GetAlertingChannelTypes(displayName string) ([]AlertingChannelType, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/api/alerting/channeltypes", nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	alertingChannelTypes := []AlertingChannelType{}
	err = json.Unmarshal(body, &alertingChannelTypes)
	if err != nil {
		return nil, err
	}

	if displayName != "" {
		filteredAlertingChannelTypes := []AlertingChannelType{}
		for _, alertingChannelType := range alertingChannelTypes {
			if alertingChannelType.DisplayName == displayName {
				filteredAlertingChannelTypes = append(filteredAlertingChannelTypes, alertingChannelType)
			}
		}

		if len(filteredAlertingChannelTypes) == 0 {
			return nil, fmt.Errorf("No alerting channel types found with display name: %s", displayName)
		}

		return filteredAlertingChannelTypes, nil
	}

	return alertingChannelTypes, nil
}
