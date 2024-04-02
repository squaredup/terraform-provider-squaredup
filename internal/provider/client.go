package provider

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

type SquaredUpClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func NewSquaredUpClient(region string, apiKey string) (*SquaredUpClient, error) {
	baseURL, err := determineBaseURL(region)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", baseURL+"/api/plugins/latest", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	client := &http.Client{}

	squaredUpClient := &SquaredUpClient{
		baseURL:    baseURL,
		apiKey:     apiKey,
		httpClient: client,
	}

	_, err = squaredUpClient.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("invalid api key with the provided region. check the api key and region and try again")
	}

	return squaredUpClient, nil
}

func determineBaseURL(region string) (string, error) {
	if region == "us" {
		return "https://api.squaredup.com", nil
	} else if region == "eu" {
		return "https://eu.api.squaredup.com", nil
	} else if strings.HasPrefix(region, "https://") {
		region = strings.TrimSuffix(region, "/")
		return region, nil
	}
	return "", fmt.Errorf("unsupported region or URL scheme: %s", region)
}

func (c *SquaredUpClient) doRequest(req *http.Request) ([]byte, error) {
	q := req.URL.Query()
	q.Add("apiKey", c.apiKey)
	req.URL.RawQuery = q.Encode()

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 && res.StatusCode != 201 && res.StatusCode != 204 {
		return nil, fmt.Errorf("status: %d, body: %s", res.StatusCode, body)
	}

	return body, err

}
