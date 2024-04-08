package provider

import (
	"fmt"
	"io"
	"net/http"
)

type SquaredUpClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	version    string
}

func NewSquaredUpClient(region string, apiKey string, version string) (*SquaredUpClient, error) {
	baseURL, err := determineBaseURL(region)
	if err != nil {
		return nil, err
	}

	return &SquaredUpClient{
		baseURL:    baseURL,
		apiKey:     apiKey,
		httpClient: &http.Client{},
		version:    version,
	}, nil
}

func determineBaseURL(region string) (string, error) {
	switch region {
	case "us":
		return "https://api.squaredup.com", nil
	case "eu":
		return "https://eu.api.squaredup.com", nil
	default:
		return "", fmt.Errorf("unsupported region: %s", region)
	}
}

func (c *SquaredUpClient) doRequest(req *http.Request) ([]byte, error) {
	q := req.URL.Query()
	q.Add("apiKey", c.apiKey)
	req.URL.RawQuery = q.Encode()

	req.Header.Add("User-Agent", fmt.Sprintf("SquaredUp-Terraform-Provider/%s", c.version))

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

	return body, nil
}
