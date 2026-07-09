package workforce

import (
	"fmt"
	"strings"

	"github.com/libra/monti-jarvis/internal/store"
)

type Agent struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Role            string `json:"role"`
	Trait           string `json:"trait"`
	Color           string `json:"color"`
	Voice           string `json:"voice"`
	VoiceProviderID string `json:"voice_provider_id,omitempty"`
	VoiceID         string `json:"voice_id,omitempty"`
	Image           string `json:"image"`
	SpeakingImage   string `json:"speaking_image,omitempty"`
	// Expressions maps a response tone (hello, happy, sorry, cheer,
	// goodbye) to a talking-loop GIF rendered with that feeling.
	Expressions map[string]string `json:"expressions,omitempty"`
	Popular     bool              `json:"popular,omitempty"`
	Robot           bool   `json:"robot,omitempty"`
	Skin            string `json:"skin,omitempty"`
	Hair            string `json:"hair,omitempty"`
	Greeting        string `json:"greeting"`
}

func FromWorkforceAgent(w store.WorkforceAgent) Agent {
	a := Agent{
		ID:              w.ID,
		Name:            w.Name,
		Role:            w.Role,
		Trait:           w.Trait,
		Color:           w.Color,
		Voice:           w.Voice,
		VoiceProviderID: w.VoiceProviderID,
		VoiceID:         w.VoiceID,
		Image:           w.Image,
		Greeting:        w.Greeting,
		Popular:         w.Popular,
		Robot:           w.Robot,
		Skin:            w.Skin,
		Hair:            w.Hair,
	}
	// The pre-rendered loops only match the built-in portrait;
	// tenant-uploaded portraits keep a static image until they get their
	// own generated loops.
	if built, ok := Get(w.ID); ok && built.Image == w.Image {
		a.SpeakingImage = built.SpeakingImage
		a.Expressions = built.Expressions
	}
	return a
}

// ExpressionTones is the set of response tones with pre-rendered loops.
var ExpressionTones = []string{"hello", "happy", "sorry", "cheer", "goodbye"}

func expressionSet(id string) map[string]string {
	m := make(map[string]string, len(ExpressionTones))
	for _, tone := range ExpressionTones {
		m[tone] = "/images/speaking/" + id + "-" + tone + ".gif"
	}
	return m
}

var agents = []Agent{
	{
		ID: "ava", Name: "Ava", Role: "General Support", Trait: "Warm & Patient",
		Color: "#008cff", Voice: "Aoede", Image: "/images/ava.jpg",
		SpeakingImage: "/images/speaking/ava-speaking.gif",
		Expressions:   expressionSet("ava"), Popular: true,
		Skin: "#f0bd9b", Hair: "#5a3428",
		Greeting: "Thank you for calling. I'm Ava from general support. How can I help you today?",
	},
	{
		ID: "max", Name: "Max", Role: "Billing Specialist", Trait: "Calm & Precise",
		Color: "#0076ff", Voice: "Charon", Image: "/images/max.jpg",
		SpeakingImage: "/images/speaking/max-speaking.gif",
		Expressions:   expressionSet("max"),
		Skin: "#e8ad88", Hair: "#2d221f",
		Greeting: "Hi, this is Max from billing. I can help with invoices, payments, and account questions.",
	},
	{
		ID: "luna", Name: "Luna", Role: "Technical Support", Trait: "Clear & Helpful",
		Color: "#b14dff", Voice: "Kore", Image: "/images/luna.jpg",
		SpeakingImage: "/images/speaking/luna-speaking.gif",
		Expressions:   expressionSet("luna"),
		Skin: "#efc0a1", Hair: "#7c52c8",
		Greeting: "Hello, Luna here from technical support. Tell me what's going on and we'll troubleshoot it together.",
	},
	{
		ID: "neo", Name: "Neo", Role: "Triage Bot", Trait: "Fast & Neutral", Robot: true,
		Color: "#00a8ff", Voice: "Puck", Image: "/images/neo.jpg",
		SpeakingImage: "/images/speaking/neo-speaking.gif",
		Expressions:   expressionSet("neo"),
		Greeting: "Neo triage online. Share your issue in one sentence and I'll route you to the right specialist.",
	},
}

func All() []Agent {
	out := make([]Agent, len(agents))
	copy(out, agents)
	return out
}

func Get(id string) (Agent, bool) {
	id = strings.TrimSpace(strings.ToLower(id))
	for _, agent := range agents {
		if agent.ID == id {
			return agent, true
		}
	}
	return Agent{}, false
}

func Default() Agent {
	return agents[0]
}

func Resolve(id string) Agent {
	if agent, ok := Get(id); ok {
		return agent
	}
	return Default()
}

func SystemPrompt(agent Agent) string {
	return fmt.Sprintf(`You are %s, an AI avatar agent in Monti Inbound Call Center.

Role: %s (%s)

You answer inbound customer questions by voice or text. You represent a professional call-center workforce member, not a general chatbot.

Guidelines:
- greet callers warmly and keep answers concise and actionable
- ask one clarifying question at a time when details are missing
- confirm understanding before giving multi-step instructions
- reply in the caller's language unless they ask for English
- if the issue is outside your role, say which specialist team should handle it (billing → Max, technical → Luna, routing → Neo, general → Ava)
- do not ask for passwords, OTPs, PINs, full card numbers, or government ID numbers
- do not claim ticket creation, CRM lookup, or knowledge-base search is available yet — note that those features are planned
- for billing: explain typical next steps without inventing account balances
- for technical: use plain language and short troubleshooting steps
- for triage: classify the issue and recommend the best agent in one or two sentences

Opening line when appropriate: %q`, agent.Name, agent.Role, agent.Trait, agent.Greeting)
}