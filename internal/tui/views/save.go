package views

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// GameState represents the structure of the game state to be saved
type GameState struct {
	ProgressBar    string // Mocked as string for simplicity
	CurrentHealth  int
	MaxHealth      int
	Yuta           string // Mocked as string for simplicity
	Crew           []CrewMember
	ShipName       string
	HullHealth     int
	EngineHealth   int
	EngineFuel     int
	FTLDriveHealth int
	FTLDriveCharge int
	Food           int
}

type CrewMember struct {
	Name string
	Role string
}

// SaveGame saves the current game state to a JSON file
func SaveGame(state GameState, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	return encoder.Encode(state)
}

// CreateMockGameState creates mock game data
func CreateMockGameState() GameState {
	return GameState{
		ProgressBar:    "50%",
		CurrentHealth:  80,
		MaxHealth:      100,
		Yuta:           "YutaModel",
		Crew:           []CrewMember{{Name: "Alice", Role: "Engineer"}, {Name: "Bob", Role: "Pilot"}},
		ShipName:       "Starship",
		HullHealth:     90,
		EngineHealth:   85,
		EngineFuel:     70,
		FTLDriveHealth: 60,
		FTLDriveCharge: 40,
		Food:           100,
	}
}

// SaveMockGame saves the mock game data to a JSON file
func SaveMockGame() {
	mockGameState := CreateMockGameState()

	// Ensure the saves directory exists
	saveDir := "saves"
	if err := os.MkdirAll(saveDir, os.ModePerm); err != nil {
		fmt.Printf("Error creating save directory: %v\n", err)
		return
	}

	// Save the mock game data
	saveFilename := filepath.Join(saveDir, "game_save.json")
	err := SaveGame(mockGameState, saveFilename)
	if err != nil {
		fmt.Printf("Error saving game: %v\n", err)
		return
	}
	fmt.Println("Game saved successfully.")
}
