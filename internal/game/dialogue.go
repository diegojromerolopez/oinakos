package game

import (
	"math/rand"
	"time"
)

type LogCategory int

const (
	LogInfo LogCategory = iota
	LogPlayer
	LogNPC
	LogCombatDamage
	LogCombatRecovery
)

type LogEntry struct {
	Text     string
	Ticks    int
	Category LogCategory
}

type DialogueEffect struct {
	Type  string `yaml:"type"`
	Value string `yaml:"value"`
}

type Choice struct {
	Text    string           `yaml:"text"`
	Next    string           `yaml:"next"`
	Effects []DialogueEffect `yaml:"effects,omitempty"`
}

type DialogueNode struct {
	Text    string   `yaml:"text"`
	Choices []Choice `yaml:"choices"`
}

type StartScenario struct {
	Weight         float64  `yaml:"weight"`
	Text           string   `yaml:"text"`
	Next           string   `yaml:"next,omitempty"`
	Choices        []Choice `yaml:"choices,omitempty"`
	AutoInitiate   bool     `yaml:"auto_initiate,omitempty"`
	ProximityRange float64  `yaml:"proximity_range,omitempty"`
}

type DialogueRoot struct {
	PlayerGreetings []string                 `yaml:"player_greetings,omitempty"`
	CombatBarks     []string                 `yaml:"combat_barks,omitempty"`
	StartScenarios  []StartScenario         `yaml:"start_scenarios"`
	Nodes           map[string]*DialogueNode `yaml:",inline"`
}

type DialogueUIState int

const (
	DialogueMinimized DialogueUIState = iota
	DialogueMaximized
)

type DialogueState struct {
	SpeakerNPC     *NPC
	CurrentText    string
	Choices        []Choice
	SelectedChoice int
	UIState        DialogueUIState
	IsActive       bool
}

func (dr *DialogueRoot) PickStart() *StartScenario {
	if len(dr.StartScenarios) == 0 {
		return nil
	}
	totalWeight := 0.0
	for _, s := range dr.StartScenarios {
		totalWeight += s.Weight
	}
	if totalWeight <= 0 {
		return &dr.StartScenarios[0]
	}

	r := rand.Float64() * totalWeight
	current := 0.0
	for i := range dr.StartScenarios {
		current += dr.StartScenarios[i].Weight
		if r <= current {
			return &dr.StartScenarios[i]
		}
	}
	return &dr.StartScenarios[0]
}

func (dr *DialogueRoot) PickGreeting() string {
	if len(dr.PlayerGreetings) == 0 {
		return "Hello."
	}
	return dr.PlayerGreetings[rand.Intn(len(dr.PlayerGreetings))]
}

func (dr *DialogueRoot) PickCombatBark() string {
	if len(dr.CombatBarks) == 0 {
		return ""
	}
	return dr.CombatBarks[rand.Intn(len(dr.CombatBarks))]
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
