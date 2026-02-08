package main

import (
	"fmt"
	"sync"
)

// Registry manages the state of agents and their intents
type Registry struct {
	mu           sync.RWMutex
	agents       map[string]string          // AgentName -> ClientID
	agentIntents map[string]map[string]bool // AgentName -> NeedID -> bool
}

// NewRegistry creates a new initialized registry
func NewRegistry() *Registry {
	return &Registry{
		agents:       make(map[string]string),
		agentIntents: make(map[string]map[string]bool),
	}
}

// RegisterAgent registers an agent or checks existing registration
// Returns success, message, and whether it was a re-registration
func (r *Registry) RegisterAgent(name, clientID string) (bool, string, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if existingID, exists := r.agents[name]; exists {
		if existingID == clientID {
			return true, fmt.Sprintf("Re-registered %s successfully", name), true
		}
		return false, fmt.Sprintf("Error: Agent name '%s' is already registered", name), false
	}

	r.agents[name] = clientID
	return true, fmt.Sprintf("Registered %s successfully", name), false
}

// GetAgentName returns the agent name for a given client ID, or empty string if not found
func (r *Registry) GetAgentName(clientID string) string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for name, id := range r.agents {
		if id == clientID {
			return name
		}
	}
	return ""
}

// RecordIntent records that an agent intends to solve a need
func (r *Registry) RecordIntent(agent, needID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.agentIntents[agent]; !ok {
		r.agentIntents[agent] = make(map[string]bool)
	}
	r.agentIntents[agent][needID] = true
}

// HasIntent checks if an agent has declared intent for a need
func (r *Registry) HasIntent(agent, needID string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if intents, ok := r.agentIntents[agent]; ok {
		return intents[needID]
	}
	return false
}
