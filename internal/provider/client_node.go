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

func (c *SquaredUpClient) GetNodes(dataSourceId string, nodeName string, nodeSourceId string, allowNull bool) ([]GremlinQueryResult, error) {
	var gremlinQueryResults []GremlinQueryResult
	var errMessage string
	for attempt := 1; attempt <= maxRetries; attempt++ {
		errMessage = fmt.Sprintf("no nodes found with name: %s in data source: %s. attempted to search for it %d times", nodeName, dataSourceId, attempt)
		rb := map[string]interface{}{
			"gremlinQuery": "g.V().has('__configId', '" + dataSourceId + "').has('name', '" + nodeName + "').hasNot('__canonicalType').valueMap(true)",
		}

		if nodeName == "" {
			rb = map[string]interface{}{
				"gremlinQuery": "g.V().has('__configId', '" + dataSourceId + "').hasNot('__canonicalType').valueMap(true)",
			}
			errMessage = fmt.Sprintf("failed to get nodes from data source: %s. attempted to search for it %d times", dataSourceId, attempt)
		}

		if nodeSourceId != "" {
			rb = map[string]interface{}{
				"gremlinQuery": "g.V().has('__configId', '" + dataSourceId + "').has('sourceId', '" + nodeSourceId + "').hasNot('__canonicalType').valueMap(true)",
			}
			errMessage = fmt.Sprintf("no nodes found with source id: %s in data source: %s. attempted to search for it %d times", nodeSourceId, dataSourceId, attempt)
		}

		reqBody, err := json.Marshal(rb)
		if err != nil {
			return nil, err
		}

		req, err := http.NewRequest("POST", c.baseURL+"/api/graph/query", strings.NewReader(string(reqBody)))
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
			if !allowNull {
				return nil, fmt.Errorf("error: %s", errMessage)
			}
		}

		gremlinQueryResults = response.GremlinQueryResults
		break

	}

	return gremlinQueryResults, nil
}
