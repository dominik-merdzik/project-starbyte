package data

import (
	"bytes"
	"encoding/json"
	"math/rand"

	_ "embed"
)

//go:embed events.json
var embeddedEvents []byte

type Event struct {
	ID          int      `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Dialogue    []string `json:"dialogue"`
	Choices     []Choice `json:"choices"`
}

type Choice struct {
	Text    string         `json:"text"`
	Effects map[string]int `json:"effects"`
	Outcome string         `json:"outcome"`
}

var Events []Event

// Load events from JSON
func LoadEvents() error {
	var data []Event

	// Decode directly from embedded JSON
	err := json.NewDecoder(bytes.NewReader(embeddedEvents)).Decode(&data)
	if err != nil {
		return err
	}

	Events = data
	return nil
}

// Pick random event
func GetRandomEvent() *Event {
	if len(Events) == 0 {
		return nil
	}

	event := Events[rand.Intn(len(Events))]
	return &event
}
