package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadConfig_AgentOrganizationAbsentKeepsExistingAgentLoading(t *testing.T) {
	configPath := writeAgentOrganizationConfig(t, `{
		"version": 3,
		"agents": {
			"list": [
				{ "id": "main", "default": true, "name": "Main" },
				{ "id": "ops", "name": "Operations" }
			]
		}
	}`)

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig() error: %v", err)
	}
	if cfg.Agents.Organization != nil {
		t.Fatalf("Agents.Organization = %+v, want nil", cfg.Agents.Organization)
	}
	if len(cfg.Agents.List) != 2 {
		t.Fatalf("Agents.List len = %d, want 2", len(cfg.Agents.List))
	}
}

func TestLoadConfig_AgentOrganizationParsesMultipleRootsAndStableSiblingOrder(t *testing.T) {
	configPath := writeAgentOrganizationConfig(t, `{
		"version": 3,
		"agents": {
			"list": [
				{ "id": "ceo", "name": "CEO" },
				{ "id": "cro", "name": "CRO" },
				{ "id": "cto", "name": "CTO" },
				{ "id": "ae", "name": "AE" },
				{ "id": "sdr", "name": "SDR" },
				{ "id": "platform", "name": "Platform" }
			],
			"organization": {
				"roots": ["ceo", "platform"],
				"nodes": [
					{ "agent_id": "sdr", "parent_agent_id": "cro", "label": "SDR Team", "group": "sales", "sort": 20 },
					{ "agent_id": "ceo", "label": "Executive", "group": "executive", "sort": 10 },
					{ "agent_id": "ae", "parent_agent_id": "cro", "label": "AE Team", "group": "sales", "sort": 10 },
					{ "agent_id": "cro", "parent_agent_id": "ceo", "label": "Revenue", "group": "executive", "sort": 20 },
					{ "agent_id": "cto", "parent_agent_id": "ceo", "label": "Technology", "group": "executive", "sort": 20 },
					{ "agent_id": "platform", "label": "Platform", "group": "shared", "sort": 30 }
				]
			}
		}
	}`)

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig() error: %v", err)
	}

	org := cfg.Agents.Organization
	if org == nil {
		t.Fatal("Agents.Organization is nil")
	}
	if got, want := org.RootAgentIDs(), []string{"ceo", "platform"}; strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("RootAgentIDs() = %v, want %v", got, want)
	}
	children := org.ChildrenOf("ceo")
	if got, want := nodeAgentIDs(children), []string{"cro", "cto"}; strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("ChildrenOf(ceo) = %v, want %v", got, want)
	}
	revenueChildren := org.ChildrenOf("cro")
	if got, want := nodeAgentIDs(revenueChildren), []string{"ae", "sdr"}; strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("ChildrenOf(cro) = %v, want %v", got, want)
	}
	if revenueChildren[0].Label != "AE Team" || revenueChildren[0].Group != "sales" {
		t.Fatalf("first revenue child metadata = %+v", revenueChildren[0])
	}
}

func TestAgentOrganizationRootAgentIDsDerivesRootsDeterministically(t *testing.T) {
	org := &AgentOrganizationConfig{
		Nodes: []AgentOrganizationNodeConfig{
			{AgentID: "beta", Sort: 10},
			{AgentID: "alpha", Sort: 10},
			{AgentID: "child", ParentAgentID: "alpha", Sort: 1},
			{AgentID: "gamma", Sort: 5},
		},
	}

	if got, want := org.RootAgentIDs(), []string{"gamma", "alpha", "beta"}; strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("RootAgentIDs() = %v, want %v", got, want)
	}
}

func TestLoadConfig_AgentOrganizationInvalidHierarchyCases(t *testing.T) {
	tests := []struct {
		name    string
		orgJSON string
		wantErr string
	}{
		{
			name: "duplicate node",
			orgJSON: `"nodes": [
				{ "agent_id": "main" },
				{ "agent_id": "main" }
			]`,
			wantErr: `agents.organization.nodes[1].agent_id "main" duplicates nodes[0]`,
		},
		{
			name: "unknown node agent",
			orgJSON: `"nodes": [
				{ "agent_id": "unknown" }
			]`,
			wantErr: `agents.organization.nodes[0].agent_id "unknown" is not a configured agent`,
		},
		{
			name: "unknown parent",
			orgJSON: `"nodes": [
				{ "agent_id": "main", "parent_agent_id": "missing" }
			]`,
			wantErr: `agents.organization.nodes[0].parent_agent_id "missing" is not a configured agent`,
		},
		{
			name: "unknown root",
			orgJSON: `"roots": ["missing"], "nodes": [
				{ "agent_id": "main" }
			]`,
			wantErr: `agents.organization.roots[0] "missing" is not a configured agent`,
		},
		{
			name: "cycle",
			orgJSON: `"nodes": [
				{ "agent_id": "main", "parent_agent_id": "ops" },
				{ "agent_id": "ops", "parent_agent_id": "main" }
			]`,
			wantErr: `agents.organization contains reporting cycle: main -> ops -> main`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configPath := writeAgentOrganizationConfig(t, `{
				"version": 3,
				"agents": {
					"list": [
						{ "id": "main", "default": true },
						{ "id": "ops" }
					],
					"organization": {
						`+tt.orgJSON+`
					}
				}
			}`)

			_, err := LoadConfig(configPath)
			if err == nil {
				t.Fatal("LoadConfig() error = nil, want invalid hierarchy error")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("LoadConfig() error = %q, want substring %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func writeAgentOrganizationConfig(t *testing.T, raw string) string {
	t.Helper()
	configPath := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(configPath, []byte(raw), 0o600); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}
	return configPath
}

func nodeAgentIDs(nodes []AgentOrganizationNodeConfig) []string {
	ids := make([]string, 0, len(nodes))
	for _, node := range nodes {
		ids = append(ids, node.AgentID)
	}
	return ids
}
