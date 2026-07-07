package scope

import "strings"

// AgentScopes maps workforce agent IDs to KM scope tags used for retrieval.
var AgentScopes = map[string][]string{
	"ava": {"general"},
	"max": {"billing"},
	"luna": {"technical"},
	"neo": {"general", "billing", "technical"},
}

// TopicScopes maps caller desk topic tabs to KM scope tags.
var TopicScopes = map[string][]string{
	"general":    {"general"},
	"billing":    {"billing"},
	"technical":  {"technical"},
}

func Resolve(agentID, topic string) []string {
	agentID = strings.ToLower(strings.TrimSpace(agentID))
	topic = strings.ToLower(strings.TrimSpace(topic))
	if topic == "" {
		topic = "general"
	}

	allowed := make(map[string]struct{})
	for _, s := range AgentScopes[agentID] {
		allowed[s] = struct{}{}
	}
	if len(allowed) == 0 {
		for _, s := range AgentScopes["ava"] {
			allowed[s] = struct{}{}
		}
	}

	var out []string
	for _, s := range TopicScopes[topic] {
		if _, ok := allowed[s]; ok {
			out = append(out, s)
		}
	}
	if len(out) == 0 {
		for s := range allowed {
			out = append(out, s)
		}
	}
	return out
}

func DefaultScope(agentID string) string {
	scopes := AgentScopes[strings.ToLower(strings.TrimSpace(agentID))]
	if len(scopes) == 0 {
		return "general"
	}
	return scopes[0]
}

func ValidAgent(agentID string) bool {
	_, ok := AgentScopes[strings.ToLower(strings.TrimSpace(agentID))]
	return ok
}