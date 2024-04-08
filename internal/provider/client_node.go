package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const maxRetries = 10
const retryDelaySeconds = 30

func (c *SquaredUpClient) GetNodes(dataSourceId string, nodeName string) ([]GremlinQueryResult, error) {
	for attempt := 1; attempt <= maxRetries; attempt++ {
		rb := map[string]interface{}{
			"gremlinQuery": "g.V().has('__configId', '" + dataSourceId + "').has('name', '" + nodeName + "').hasNot('__canonicalType').order().valueMap(true)",
		}

		if nodeName == "" {
			rb = map[string]interface{}{
				"gremlinQuery": "g.V().has('__configId', '" + dataSourceId + "').hasNot('__canonicalType').order().valueMap(true)",
			}
		}

		reqBody, err := json.Marshal(rb)
		if err != nil {
			return nil, err
		}

		req, err := http.NewRequest("POST", c.baseURL+"/api/query", strings.NewReader(string(reqBody)))
		if err != nil {
			return nil, err
		}

		body, err := c.doRequest(req)
		if err != nil {
			if attempt < maxRetries {
				time.Sleep(retryDelaySeconds * time.Second)
				continue
			}
			return nil, err
		}

		var response SquaredupGremlinQuery
		err = json.Unmarshal(body, &response)
		if err != nil {
			return nil, err
		}

		if len(response.GremlinQueryResults) == 0 {
			if attempt < maxRetries {
				time.Sleep(retryDelaySeconds * time.Second)
				continue
			}
			return nil, fmt.Errorf("no nodes found with name: %s in Data Source: %s. attempted to search for it %d times", nodeName, dataSourceId, attempt)
		}

		return response.GremlinQueryResults, nil
	}

	return nil, nil
}
