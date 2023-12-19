package provider

import "github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"

type LatestDataSource struct {
	Category    string `json:"category"`
	Description string `json:"description"`
	Author      string `json:"author"`
	LastUpdated string `json:"lastUpdated"`
	LambdaName  string `json:"lambdaName"`
	Version     string `json:"version"`
	OnPrem      bool   `json:"onPrem"`
	DisplayName string `json:"displayName"`
	PluginID    string `json:"id"`
}

type DataSource struct {
	DisplayName string `json:"displayName"`
	ID          string `json:"id,omitempty"`
	Plugin      struct {
		Name string `json:"name"`
	} `json:"plugin"`
	AgentGroupID string `json:"agentGroupId,omitempty"`
}

type WorkspaceConfig struct {
	DisplayName string              `json:"displayName"`
	ID          string              `json:"id,omitempty"`
	Links       WorkspaceLinks      `json:"links"`
	Properties  WorkspaceProperties `json:"properties"`
}

type WorkspaceLinks struct {
	Plugins    []string `json:"plugins"`
	Workspaces []string `json:"workspaces"`
}

type WorkspaceProperties struct {
	OpenAccessEnabled bool     `json:"openAccessEnabled"`
	Tags              []string `json:"tags"`
	Description       string   `json:"description"`
	Type              string   `json:"type,omitempty"`
}

type WorkspaceRead struct {
	ID          string            `json:"id"`
	Type        string            `json:"type"`
	DisplayName string            `json:"displayName"`
	Tenant      string            `json:"tenant"`
	ConfigID    string            `json:"configId"`
	Data        WorkspaceReadData `json:"data"`
}

type WorkspaceReadData struct {
	ID            string              `json:"id"`
	Label         string              `json:"label"`
	LinkedObjects string              `json:"linkedObjects"`
	Properties    WorkspaceProperties `json:"properties"`
	SourceType    string              `json:"sourceType"`
	SourceName    string              `json:"sourceName"`
	Search        string              `json:"__search"`
	Name          string              `json:"__name"`
	PartitionKey  string              `json:"__partitionKey"`
	Links         WorkspaceLinks      `json:"links"`
}

type DataSourceDataStreams struct {
	DisplayName         string `json:"displayName"`
	DataSourceName      string `json:"dataSourceName"`
	LastUpdated         string `json:"lastUpdated"`
	ParentPluginVersion string `json:"parentPluginVersion"`
	ParentPluginID      string `json:"parentPluginId"`
	Type                string `json:"type"`
	ID                  string `json:"id"`
	Definition          struct {
		Name string `json:"name"`
	} `json:"definition"`
}

type Dashboard struct {
	DisplayName   string               `json:"displayName"`
	LastUpdated   string               `json:"lastUpdated"`
	WorkspaceID   string               `json:"workspaceId"`
	ID            string               `json:"id"`
	Content       jsontypes.Normalized `json:"content"`
	Group         string               `json:"group,omitempty"`
	Name          string               `json:"name"`
	SchemaVersion string               `json:"schemaVersion"`
	Timeframe     string               `json:"timeframe,omitempty"`
}

type SquaredupGremlinQuery struct {
	GremlinQueryResults []GremlinQueryResult `json:"gremlinQueryResults"`
}

type GremlinQueryResult struct {
	ID           string   `json:"id"`
	Label        string   `json:"label"`
	SourceName   []string `json:"sourceName"`
	Type         []string `json:"type"`
	SourceType   []string `json:"sourceType"`
	Name         []string `json:"name"`
	SourceId     []string `json:"sourceId"`
	Search       []string `json:"__search"`
	DisplayName  []string `json:"__name"`
	PartitionKey []string `json:"__partitionKey"`
	TenantId     []string `json:"__tenantId"`
	ConfigId     []string `json:"__configId"`
}
