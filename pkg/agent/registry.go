package agent

import (
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/sipeed/picoclaw/pkg/bus"
	"github.com/sipeed/picoclaw/pkg/config"
	"github.com/sipeed/picoclaw/pkg/logger"
	"github.com/sipeed/picoclaw/pkg/providers"
	"github.com/sipeed/picoclaw/pkg/routing"
	"github.com/sipeed/picoclaw/pkg/tools"
)

// AgentDescriptor is the compact discovery model exposed to peer agents.
type AgentDescriptor struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// AgentRegistry manages multiple agent instances and routes messages to them.
type AgentRegistry struct {
	agents   map[string]*AgentInstance
	resolver *routing.RouteResolver
	mu       sync.RWMutex
}

// NewAgentRegistry creates a registry from config, instantiating all agents.
func NewAgentRegistry(
	cfg *config.Config,
	provider providers.LLMProvider,
) *AgentRegistry {
	registry := &AgentRegistry{
		agents:   make(map[string]*AgentInstance),
		resolver: routing.NewRouteResolver(cfg),
	}

	agentConfigs := cfg.Agents.List
	if len(agentConfigs) == 0 {
		implicitAgent := &config.AgentConfig{
			ID:      "main",
			Default: true,
		}
		instance := NewAgentInstance(implicitAgent, &cfg.Agents.Defaults, cfg, provider)
		registry.agents["main"] = instance
		logger.InfoCF("agent", "Created implicit main agent (no agents.list configured)", nil)
	} else {
		for i := range agentConfigs {
			ac := &agentConfigs[i]
			id := routing.NormalizeAgentID(ac.ID)
			instance := NewAgentInstance(ac, &cfg.Agents.Defaults, cfg, provider)
			registry.agents[id] = instance
			logger.InfoCF("agent", "Registered agent",
				map[string]any{
					"agent_id":  id,
					"name":      ac.Name,
					"workspace": instance.Workspace,
					"model":     instance.Model,
				})
		}
	}

	registry.installAgentDiscoveryPrompts()

	return registry
}

// GetAgent returns the agent instance for a given ID.
func (r *AgentRegistry) GetAgent(agentID string) (*AgentInstance, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	id := routing.NormalizeAgentID(agentID)
	agent, ok := r.agents[id]
	return agent, ok
}

// ResolveRoute determines which agent handles the normalized inbound context.
func (r *AgentRegistry) ResolveRoute(inbound bus.InboundContext) routing.ResolvedRoute {
	return r.resolver.ResolveRoute(inbound)
}

// ListAgentIDs returns all registered agent IDs.
func (r *AgentRegistry) ListAgentIDs() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ids := make([]string, 0, len(r.agents))
	for id := range r.agents {
		ids = append(ids, id)
	}
	slices.Sort(ids)
	return ids
}

// ListAgentDescriptors returns compact descriptors for all registered agents.
func (r *AgentRegistry) ListAgentDescriptors() []AgentDescriptor {
	r.mu.RLock()
	defer r.mu.RUnlock()

	descriptors := make([]AgentDescriptor, 0, len(r.agents))
	for _, agent := range r.agents {
		descriptors = append(descriptors, agentDescriptor(agent))
	}
	slices.SortFunc(descriptors, func(a, b AgentDescriptor) int {
		return strings.Compare(a.ID, b.ID)
	})
	return descriptors
}

// GetAgentDescriptor returns the discovery descriptor for a normalized agent ID.
func (r *AgentRegistry) GetAgentDescriptor(agentID string) (AgentDescriptor, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	id := routing.NormalizeAgentID(agentID)
	agent, ok := r.agents[id]
	if !ok {
		return AgentDescriptor{}, false
	}
	return agentDescriptor(agent), true
}

func (r *AgentRegistry) installAgentDiscoveryPrompts() {
	descriptors := r.ListAgentDescriptors()
	if len(descriptors) <= 1 {
		return
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, agent := range r.agents {
		if agent == nil || agent.ContextBuilder == nil {
			continue
		}
		if err := agent.ContextBuilder.RegisterPromptContributor(agentDiscoveryPromptContributor{
			selfID:      agent.ID,
			descriptors: descriptors,
		}); err != nil {
			logger.WarnCF("agent", "Failed to register agent discovery prompt contributor", map[string]any{
				"agent_id": agent.ID,
				"error":    err.Error(),
			})
		}
	}
}

func agentDescriptor(agent *AgentInstance) AgentDescriptor {
	if agent == nil {
		return AgentDescriptor{}
	}

	name := strings.TrimSpace(agent.Name)
	description := strings.TrimSpace("Workspace: " + filepath.Base(agent.Workspace))
	if agent.ContextBuilder != nil {
		definition := agent.ContextBuilder.LoadAgentDefinition()
		if definition.Agent != nil {
			frontmatter := definition.Agent.Frontmatter
			if strings.TrimSpace(frontmatter.Name) != "" {
				name = strings.TrimSpace(frontmatter.Name)
			}
			if strings.TrimSpace(frontmatter.Description) != "" {
				description = strings.TrimSpace(frontmatter.Description)
			}
		}
	}
	if name == "" {
		name = agent.ID
	}
	if description == "" || description == "Workspace: ." {
		description = "Agent " + agent.ID
	}

	return AgentDescriptor{
		ID:          agent.ID,
		Name:        compactDescriptorText(name),
		Description: compactDescriptorText(description),
	}
}

func compactDescriptorText(value string) string {
	const maxLen = 240

	value = strings.Join(strings.Fields(value), " ")
	runes := []rune(value)
	if len(runes) <= maxLen {
		return value
	}
	return strings.TrimSpace(string(runes[:maxLen]))
}

// CanSpawnSubagent checks if parentAgentID is allowed to spawn targetAgentID.
func (r *AgentRegistry) CanSpawnSubagent(parentAgentID, targetAgentID string) bool {
	parent, ok := r.GetAgent(parentAgentID)
	if !ok {
		return false
	}
	if parent.Subagents == nil || parent.Subagents.AllowAgents == nil {
		return false
	}
	targetNorm := routing.NormalizeAgentID(targetAgentID)
	for _, allowed := range parent.Subagents.AllowAgents {
		if allowed == "*" {
			return true
		}
		if routing.NormalizeAgentID(allowed) == targetNorm {
			return true
		}
	}
	return false
}

// ForEachTool calls fn for every tool registered under the given name
// across all agents. This is useful for propagating dependencies (e.g.
// MediaStore) to tools after registry construction.
func (r *AgentRegistry) ForEachTool(name string, fn func(tools.Tool)) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, agent := range r.agents {
		if t, ok := agent.Tools.Get(name); ok {
			fn(t)
		}
	}
}

// Close releases resources held by all registered agents.
func (r *AgentRegistry) Close() {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, agent := range r.agents {
		if err := agent.Close(); err != nil {
			logger.WarnCF("agent", "Failed to close agent",
				map[string]any{"agent_id": agent.ID, "error": err.Error()})
		}
	}
}

// GetDefaultAgent returns the default agent instance.
func (r *AgentRegistry) GetDefaultAgent() *AgentInstance {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, agent := range r.agents {
		if agent != nil && agent.Default {
			return agent
		}
	}
	if agent, ok := r.agents["main"]; ok {
		return agent
	}
	for _, agent := range r.agents {
		return agent
	}
	return nil
}
