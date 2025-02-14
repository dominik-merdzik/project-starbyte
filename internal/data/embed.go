package data

import (
	_ "embed"
	"encoding/json"
)

// MissionsFile is the top-level structure for the missions JSON data
type MissionsFile struct {
	Main     []MainMission     `json:"main"`
	Received []ReceivedMission `json:"received"`
}

// MainMission represents a main mission (story missions)
type MainMission struct {
	Step         int    `json:"Step"`
	Title        string `json:"Title"`
	Description  string `json:"Description"`
	Status       string `json:"Status"`
	Location     string `json:"Location"`
	Income       string `json:"Income"`
	Requirements string `json:"Requirements"`
	Received     string `json:"Received"`
	Category     string `json:"Category"`
}

// ReceivedMission represents a group of missions available at a specific location (this will neeed to chanage later)
type ReceivedMission struct {
	Location string `json:"Location"`
	NPCs     []NPC  `json:"NPCs"`
}

// NPC represents a non-player character offering missions
type NPC struct {
	Name     string    `json:"Name"`
	Missions []Mission `json:"Missions"`
}

// Mission represents an individual mission offered by an NPC
type Mission struct {
	Title        string `json:"Title"`
	Description  string `json:"Description"`
	Status       string `json:"Status"`
	Location     string `json:"Location"`
	Income       string `json:"Income"`
	Requirements string `json:"Requirements"`
	Received     string `json:"Received"`
	Category     string `json:"Category"`
}

//go:embed missions.json
var missionsJSON []byte

// LoadMissions loads and unmarshals the embedded missions JSON
func LoadMissions() (MissionsFile, error) {
	var mf MissionsFile
	err := json.Unmarshal(missionsJSON, &mf)
	return mf, err
}
