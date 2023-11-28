package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const maxRetries = 3
const retryDelaySeconds = 5

func (c *SquaredUpClient) GetNodes(dataSourceId string, nodeName string) ([]GremlinQueryResult, error) {
	for attempt := 1; attempt <= maxRetries; attempt++ {
		rb := map[string]interface{}{
			"gremlinQuery": "g.V().has('__configId', '" + dataSourceId + "').has('name', '" + nodeName + "').valueMap(true)",
		}

		if nodeName == "" {
			rb = map[string]interface{}{
				"gremlinQuery": "g.V().has('__configId', '" + dataSourceId + "').valueMap(true)",
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
			return nil, fmt.Errorf("No nodes found with name: %s in Data Source: %s. Attempted to search for it %d times", nodeName, dataSourceId, attempt)
		}

		return response.GremlinQueryResults, nil
	}

	return nil, nil
}
