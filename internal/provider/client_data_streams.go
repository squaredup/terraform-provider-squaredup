package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (c *SquaredUpClient) GetDataStreams(dataSourceId string, DataStreamDefinitionName string) ([]DataSourceDataStreams, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/api/config/datastreams/plugin/"+dataSourceId, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	dataStreams := []DataSourceDataStreams{}
	err = json.Unmarshal(body, &dataStreams)
	if err != nil {
		return nil, err
	}

	if DataStreamDefinitionName != "" {
		filteredDataStreams := []DataSourceDataStreams{}
		for _, dataStream := range dataStreams {
			if dataStream.Definition.Name == DataStreamDefinitionName {
				filteredDataStreams = append(filteredDataStreams, dataStream)
			}
		}

		if len(filteredDataStreams) == 0 {
			return nil, fmt.Errorf("No data streams found with data source name: %s", DataStreamDefinitionName)
		}

		return filteredDataStreams, nil
	}

	return dataStreams, nil
}
