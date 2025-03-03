package data

import (
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// seed the random number generator
func init() {
	rand.Seed(time.Now().UnixNano())
}

// SaveFilePath defines where the save file will be stored - lets keep it simple for now
const SaveFilePath = "temp/save.json"

// ---------------------
// Save File Structures
// ---------------------

type FullGameSave struct {
	GameTitle    string       `json:"gameTitle"`
	GameMetadata GameMetadata `json:"gameMetadata"`
	Player       Player       `json:"player"`
	Ship         Ship         `json:"ship"`
	Crew         []CrewMember `json:"crew"`
}

type GameMetadata struct {
	Version            string             `json:"version"`
	DateCreated        string             `json:"dateCreated"`
	LastSaveTime       string             `json:"lastSaveTime"`
	TotalPlayTime      TotalPlayTime      `json:"totalPlayTime"`
	DifficultySettings DifficultySettings `json:"difficultySettings"`
}

type TotalPlayTime struct {
	Hours   int `json:"hours"`
	Minutes int `json:"minutes"`
	Seconds int `json:"seconds"`
}

type DifficultySettings struct {
	DifficultyLevel    string  `json:"difficultyLevel"`
	ResourceMultiplier float64 `json:"resourceMultiplier"`
	CrewMoraleImpact   float64 `json:"crewMoraleImpact"`
}

type Player struct {
	PlayerId         string     `json:"playerId"`
	PlayerName       string     `json:"playerName"`
	Faction          string     `json:"faction"`
	ExperiencePoints int        `json:"experiencePoints"`
	Level            int        `json:"level"`
	Credits          int        `json:"credits"`
	Reputation       Reputation `json:"reputation"`
}

type Reputation struct {
	AlliedFactions map[string]int `json:"alliedFactions"`
	EnemyFactions  map[string]int `json:"enemyFactions"`
}

type Ship struct {
	ShipId            string   `json:"shipId"`
	ShipName          string   `json:"shipName"`
	HullIntegrity     int      `json:"hullIntegrity"`
	MaxHullIntegrity  int      `json:"maxHullIntegrity"`
	ShieldStrength    int      `json:"shieldStrength"`
	MaxShieldStrength int      `json:"maxShieldStrength"`
	Fuel              int      `json:"fuel"`
	MaxFuel           int      `json:"maxFuel"`
	EngineHealth      int      `json:"engineHealth"`
	FTLDriveHealth    int      `json:"ftlDriveHealth"`
	FTLDriveCharge    int      `json:"ftlDriveCharge"`
	Food              int      `json:"food"`
	Location          Location `json:"location"`
	Cargo             Cargo    `json:"cargo"`
	Modules           []Module `json:"modules"`
	Upgrades          Upgrades `json:"upgrades"`
}

type Location struct {
	StarSystemId string      `json:"starSystemId"`
	PlanetId     string      `json:"planetId"`
	Coordinates  Coordinates `json:"coordinates"`
}

type Coordinates struct {
	X int `json:"x"`
	Y int `json:"y"`
	Z int `json:"z"`
}

type Cargo struct {
	Capacity     int         `json:"capacity"`
	UsedCapacity int         `json:"usedCapacity"`
	Items        []CargoItem `json:"items"`
}

type CargoItem struct {
	ItemId   string `json:"itemId"`
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
}

type Module struct {
	ModuleId string `json:"moduleId"`
	Name     string `json:"name"`
	Level    int    `json:"level"`
	Status   string `json:"status"`
}

type Upgrades struct {
	Engine         UpgradeLevel `json:"engine"`
	WeaponSystems  UpgradeLevel `json:"weaponSystems"`
	CargoExpansion UpgradeLevel `json:"cargoExpansion"`
}

type UpgradeLevel struct {
	CurrentLevel int `json:"currentLevel"`
	MaxLevel     int `json:"maxLevel"`
}

type CrewMember struct {
	CrewId         string  `json:"crewId"`
	Name           string  `json:"name"`
	Role           string  `json:"role"`
	Degree         int     `json:"degree"`
	Experience     int     `json:"experience"`
	Morale         int     `json:"morale"`
	Health         int     `json:"health"`
	Skills         Skills  `json:"skills"`
	AssignedTaskId *string `json:"assignedTaskId"`
}

type Skills struct {
	Piloting    int `json:"piloting"`
	Engineering int `json:"engineering"`
	Combat      int `json:"combat"`
}

// -------------------
// helper Functions
// -------------------

// generateRandomID returns a random ID string with the given prefix
func generateRandomID(prefix string) string {
	return prefix + strconv.Itoa(rand.Intn(1000000))
}

// ------------------------------
// full game save file operations
// ------------------------------

// CreateNewFullGameSave creates a new full game save file using the provided parameters
func CreateNewFullGameSave(difficulty, shipName, startingLocation string) error {
	now := time.Now()

	// Build the full game save structure with default values for a new game.
	fullSave := FullGameSave{
		GameTitle: "Project Starbyte",
		GameMetadata: GameMetadata{
			Version:      "0.0.1",
			DateCreated:  now.Format("2006-01-02"),
			LastSaveTime: now.Format(time.RFC3339),
			TotalPlayTime: TotalPlayTime{
				Hours:   0,
				Minutes: 0,
				Seconds: 0,
			},
			DifficultySettings: DifficultySettings{
				DifficultyLevel:    difficulty,
				ResourceMultiplier: 1.0,
				CrewMoraleImpact:   1.0,
			},
		},
		Player: Player{
			PlayerId:         generateRandomID("PLAYER_"),
			PlayerName:       "Commander " + shipName,
			Faction:          "Independent",
			ExperiencePoints: 0,
			Level:            1,
			Credits:          1000,
			Reputation: Reputation{
				AlliedFactions: map[string]int{
					"GalacticUnion": 50,
				},
				EnemyFactions: map[string]int{
					"PirateClan": -20,
				},
			},
		},
		Ship: Ship{
			ShipId:            generateRandomID("SHIP_"),
			ShipName:          shipName,
			HullIntegrity:     100,
			MaxHullIntegrity:  100,
			ShieldStrength:    50,
			MaxShieldStrength: 50,
			Fuel:              100,
			MaxFuel:           200,
			EngineHealth:      100, // New default.
			FTLDriveHealth:    70,  // New default.
			FTLDriveCharge:    0,   // New default.
			Food:              100, // New default.
			Location: Location{
				StarSystemId: generateRandomID("SYS_"),
				PlanetId:     startingLocation,
				Coordinates: Coordinates{
					X: 0,
					Y: 0,
					Z: 0,
				},
			},
			Cargo: Cargo{
				Capacity:     100,
				UsedCapacity: 0,
				Items: []CargoItem{
					{
						ItemId:   generateRandomID("ITEM_"),
						Name:     "Iron Ore",
						Quantity: 10,
					},
					{
						ItemId:   generateRandomID("ITEM_"),
						Name:     "Water",
						Quantity: 5,
					},
				},
			},
			Modules: []Module{
				{
					ModuleId: generateRandomID("MOD_ENG_"),
					Name:     "Basic Engine",
					Level:    1,
					Status:   "operational",
				},
				{
					ModuleId: generateRandomID("MOD_LIFE_"),
					Name:     "Life Support",
					Level:    1,
					Status:   "operational",
				},
			},
			Upgrades: Upgrades{
				Engine: UpgradeLevel{
					CurrentLevel: 1,
					MaxLevel:     5,
				},
				WeaponSystems: UpgradeLevel{
					CurrentLevel: 0,
					MaxLevel:     5,
				},
				CargoExpansion: UpgradeLevel{
					CurrentLevel: 0,
					MaxLevel:     5,
				},
			},
		},
		Crew: []CrewMember{
			{
				CrewId:         generateRandomID("CREW_"),
				Name:           "Alice",
				Role:           "Pilot",
				Degree:         1,
				Experience:     0,
				Morale:         100,
				Health:         100,
				Skills:         Skills{Piloting: 5, Engineering: 1, Combat: 2},
				AssignedTaskId: nil,
			},
			{
				CrewId:         generateRandomID("CREW_"),
				Name:           "Bob",
				Role:           "Engineer",
				Degree:         1,
				Experience:     0,
				Morale:         95,
				Health:         100,
				Skills:         Skills{Piloting: 1, Engineering: 5, Combat: 1},
				AssignedTaskId: nil,
			},
		},
	}

	// Wrap the save data in an array (slice) as per your JSON structure.
	saveData := []FullGameSave{fullSave}

	data, err := json.MarshalIndent(saveData, "", "  ")
	if err != nil {
		return err
	}

	// Ensure the directory exists.
	dir := filepath.Dir(SaveFilePath)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}

	return ioutil.WriteFile(SaveFilePath, data, 0644)
}

// DefaultFullGameSave returns a default FullGameSave structure with initial "new game" values
func DefaultFullGameSave() *FullGameSave {
	now := time.Now()
	return &FullGameSave{
		GameTitle: "Project Starbyte",
		GameMetadata: GameMetadata{
			Version:      "0.0.1",
			DateCreated:  now.Format("2006-01-02"),
			LastSaveTime: now.Format(time.RFC3339),
			TotalPlayTime: TotalPlayTime{
				Hours:   0,
				Minutes: 0,
				Seconds: 0,
			},
			DifficultySettings: DifficultySettings{
				DifficultyLevel:    "Normal",
				ResourceMultiplier: 1.0,
				CrewMoraleImpact:   1.0,
			},
		},
		Player: Player{
			PlayerId:         "PLAYER_001",
			PlayerName:       "Commander Default",
			Faction:          "Independent",
			ExperiencePoints: 0,
			Level:            1,
			Credits:          1000,
			Reputation: Reputation{
				AlliedFactions: map[string]int{
					"GalacticUnion": 50,
				},
				EnemyFactions: map[string]int{
					"PirateClan": -20,
				},
			},
		},
		Ship: Ship{
			ShipId:            "SHIP_001",
			ShipName:          "Default Ship",
			HullIntegrity:     100,
			MaxHullIntegrity:  100,
			ShieldStrength:    50,
			MaxShieldStrength: 50,
			Fuel:              100,
			MaxFuel:           200,
			EngineHealth:      100, // New default.
			FTLDriveHealth:    70,  // New default.
			FTLDriveCharge:    0,   // New default.
			Food:              100, // New default.
			Location: Location{
				StarSystemId: "SYS_0001",
				PlanetId:     "Earth",
				Coordinates: Coordinates{
					X: 0,
					Y: 0,
					Z: 0,
				},
			},
			Cargo: Cargo{
				Capacity:     100,
				UsedCapacity: 0,
				Items: []CargoItem{
					{
						ItemId:   "ITEM_001",
						Name:     "Iron Ore",
						Quantity: 10,
					},
					{
						ItemId:   "ITEM_002",
						Name:     "Water",
						Quantity: 5,
					},
				},
			},
			Modules: []Module{
				{
					ModuleId: "MOD_001",
					Name:     "Basic Engine",
					Level:    1,
					Status:   "operational",
				},
				{
					ModuleId: "MOD_002",
					Name:     "Life Support",
					Level:    1,
					Status:   "operational",
				},
			},
			Upgrades: Upgrades{
				Engine: UpgradeLevel{
					CurrentLevel: 1,
					MaxLevel:     5,
				},
				WeaponSystems: UpgradeLevel{
					CurrentLevel: 0,
					MaxLevel:     5,
				},
				CargoExpansion: UpgradeLevel{
					CurrentLevel: 0,
					MaxLevel:     5,
				},
			},
		},
		Crew: []CrewMember{
			{
				CrewId:         "CREW_001",
				Name:           "Alice",
				Role:           "Pilot",
				Degree:         1,
				Experience:     0,
				Morale:         100,
				Health:         100,
				Skills:         Skills{Piloting: 5, Engineering: 1, Combat: 2},
				AssignedTaskId: nil,
			},
			{
				CrewId:         "CREW_002",
				Name:           "Bob",
				Role:           "Engineer",
				Degree:         1,
				Experience:     0,
				Morale:         95,
				Health:         100,
				Skills:         Skills{Piloting: 1, Engineering: 5, Combat: 1},
				AssignedTaskId: nil,
			},
		},
	}
}

// SaveExists checks whether a save file already exists
func SaveExists() bool {
	_, err := os.Stat(SaveFilePath)
	return err == nil
}

// LoadFullGameSave reads the JSON save file and returns the full game data
func LoadFullGameSave() (*FullGameSave, error) {
	data, err := ioutil.ReadFile(SaveFilePath)
	if err != nil {
		return nil, err
	}
	var saves []FullGameSave
	if err := json.Unmarshal(data, &saves); err != nil {
		return nil, err
	}
	if len(saves) == 0 {
		return nil, nil
	}
	return &saves[0], nil
}
