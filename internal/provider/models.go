package provider

import "github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"

type LatestDataSource struct {
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
		Name   string `json:"name"`
		OnPrem bool   `json:"onPrem"`
	} `json:"plugin"`
	AgentGroupID string `json:"agentGroupId,omitempty"`
}

type WorkspaceRead struct {
	ID          string            `json:"id"`
	DisplayName string            `json:"displayName"`
	Data        WorkspaceReadData `json:"data"`
}

type WorkspaceReadData struct {
	AlertingRules []WorkspaceAlertData `json:"alertingRules,omitempty"`
	LinkedObjects string               `json:"linkedObjects"`
	Properties    WorkspaceProperties  `json:"properties"`
	Links         WorkspaceLinks       `json:"links"`
}

type WorkspaceProperties struct {
	DashboardSharingEnabled bool          `json:"openAccessEnabled"`
	Tags                    []string      `json:"tags"`
	Description             string        `json:"description"`
	Type                    string        `json:"type,omitempty"`
	AuthorizedEmailDomains  []string      `json:"authorizedEmailDomains"`
	DashboardIdOrder        []interface{} `json:"dashboardIdOrder,omitempty"`
}

type WorkspaceLinks struct {
	Plugins    []string `json:"plugins"`
	Workspaces []string `json:"workspaces"`
}

type DataSourceDataStreams struct {
	DisplayName    string `json:"displayName"`
	DataSourceName string `json:"dataSourceName"`
	ID             string `json:"id"`
	Definition     struct {
		Name string `json:"name"`
	} `json:"definition"`
}

type Dashboard struct {
	DisplayName   string               `json:"displayName"`
	LastUpdated   string               `json:"lastUpdated"`
	WorkspaceID   string               `json:"workspaceId"`
	ID            string               `json:"id"`
	Content       jsontypes.Normalized `json:"content"`
	SchemaVersion string               `json:"schemaVersion"`
	Timeframe     string               `json:"timeframe,omitempty"`
}

type SquaredupGremlinQuery struct {
	GremlinQueryResults []GremlinQueryResult `json:"gremlinQueryResults"`
}

type GremlinQueryResult struct {
	ID          string   `json:"id"`
	SourceName  []string `json:"sourceName"`
	DisplayName []string `json:"name"`
	SourceID    []string `json:"sourceId"`
	Type        []string `json:"type"`
}

type DashboardShare struct {
	LastUpdated string                   `json:"lastUpdated,omitempty"`
	ID          string                   `json:"id,omitempty"`
	TargetID    string                   `json:"targetId"`
	WorkspaceID string                   `json:"workspaceId"`
	Properties  DashboardShareProperties `json:"properties"`
}

type DashboardShareProperties struct {
	Enabled               bool `json:"enabled"`
	RequireAuthentication bool `json:"requireAuthentication"`
}

type AlertingChannelType struct {
	ChannelID             string `json:"id"`
	DisplayName           string `json:"displayName"`
	Protocol              string `json:"protocol"`
	ImagePreviewSupported bool   `json:"imagePreviewSupported"`
	Description           string `json:"description"`
}

type AlertingChannel struct {
	ID            string                 `json:"id"`
	DisplayName   string                 `json:"displayName"`
	Description   string                 `json:"description"`
	ChannelTypeID string                 `json:"channelTypeId"`
	Config        map[string]interface{} `json:"config"`
	Enabled       bool                   `json:"enabled"`
}

type WorkspaceAlertsData struct {
	AlertingRules []WorkspaceAlertData `json:"alertingRules"`
}

type WorkspaceAlertData struct {
	Channels   []AlertChannel  `json:"channels"`
	Conditions AlertConditions `json:"conditions"`
}

type AlertChannel struct {
	ID                  string `json:"id"`
	IncludePreviewImage bool   `json:"includePreviewImage"`
}

type AlertConditions struct {
	Monitors AlertMonitors `json:"monitors"`
}

type AlertMonitors struct {
	IncludeAllTiles       bool                      `json:"includeAllTiles"`
	DashboardRollupHealth bool                      `json:"dashboardRollupHealth"`
	RollupHealth          bool                      `json:"rollupHealth"`
	Dashboards            map[string]AlertDashboard `json:"dashboards,omitempty"`
}

type AlertDashboard struct {
	Tiles map[string]AlertTile `json:"tiles"`
}

type AlertTile struct {
	Include bool `json:"include"`
}

type Script struct {
	DisplayName string       `json:"displayName"`
	ScriptType  string       `json:"scriptType,omitempty"`
	SubType     string       `json:"subType,omitempty"`
	Config      ScriptConfig `json:"config"`
	ID          string       `json:"id,omitempty"`
}

type ScriptConfig struct {
	Src string `json:"src"`
}

type ScopeCreate struct {
	Scope Scope `json:"scope"`
}

type Scope struct {
	Name        string                 `json:"name"`
	Version     int                    `json:"version"`
	Query       string                 `json:"query"`
	Bindings    map[string]interface{} `json:"bindings,omitempty"`
	QueryDetail ScopeQueryDetail       `json:"queryDetail"`
}

type ScopeQueryDetail struct {
	IDs          []string                 `json:"ids,omitempty"`
	Plugins      []ScopeQueryDetailPlugin `json:"plugins,omitempty"`
	Types        []ScopeQueryDetailType   `json:"types,omitempty"`
	BooleanQuery string                   `json:"booleanQuery,omitempty"`
}

type ScopeQueryDetailPlugin struct {
	Value string `json:"value,omitempty"`
}

type ScopeQueryDetailType struct {
	Value string `json:"value,omitempty"`
}

type ScopeRead struct {
	ID          string    `json:"id"`
	DisplayName string    `json:"displayName"`
	Data        ScopeData `json:"data"`
	WorkspaceID string    `json:"workspaceId"`
}

type ScopeData struct {
	ID          string `json:"id"`
	Query       string `json:"query"`
	QueryDetail string `json:"queryDetail"`
}

type DashboardImage struct {
	DataURL  string                 `json:"dataURL"`
	Metadata DashboardImageMetadata `json:"metadata"`
}

type DashboardImageMetadata struct {
	FileName string `json:"fileName"`
}

type DashboardVariable struct {
	Name                   string `json:"name"`
	Type                   string `json:"type"`
	ScopeID                string `json:"scope"`
	Default                string `json:"default"`
	AllowMultipleSelection bool   `json:"allowMultipleSelection"`
	DashboardID            string `json:"dashboardId,omitempty"`
}

type DashboardVariableRead struct {
	WorkspaceID string            `json:"workspaceId"`
	Content     DashboardVariable `json:"content"`
	DisplayName string            `json:"displayName"`
	Name        string            `json:"name"`
	ID          string            `json:"id"`
}
