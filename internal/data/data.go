package data

import (
	"encoding/json"
	"io/ioutil"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

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
	Missions     Missions     `json:"missions"`
	GameMap      GameMap      `json:"gameMap"`
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

// ---------------------
// Map structures
// ---------------------
type GameMap struct {
	StarSystems []StarSystem `json:"starSystems"`
}

type StarSystem struct {
	SystemID    string      `json:"systemId"`
	Name        string      `json:"name"`
	Coordinates Coordinates `json:"coordinates"`
	Planets     []Planet    `json:"planets"`
}

type Planet struct {
	PlanetID    string      `json:"planetId"`
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Resources   []Resource  `json:"resources"`
	Coordinates Coordinates `json:"coordinates"`
}

type Resource struct {
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
}

// ---------------------
// mission structures
// ---------------------

type Mission struct {
	Step              int      `json:"Step,omitempty"`
	MissionId         string   `json:"missionId"`
	Title             string   `json:"Title"`
	Description       string   `json:"Description"`
	Status            string   `json:"Status"`
	Location          string   `json:"Location"`
	Income            int      `json:"Income"`
	Requirements      string   `json:"Requirements"`
	Received          string   `json:"Received"`
	Category          string   `json:"Category"`
	TravelTime        int      `json:"TravelTime"`
	FuelNeeded        int      `json:"FuelNeeded"`
	DestinationPlanet string   `json:"DestinationPlanet"`
	Dialogue          []string `json:"dialogue"`
}

type NPC struct {
	Name     string    `json:"Name"`
	Missions []Mission `json:"Missions"`
}

type ReceivedMissionGroup struct {
	Location string `json:"Location"`
	NPCs     []NPC  `json:"NPCs"`
}

type Missions struct {
	Main     []Mission              `json:"main"`
	Received []ReceivedMissionGroup `json:"received"`
}

// -------------------
// helper functions
// -------------------

// Searches for a planet by name across all star systems
func FindPlanet(gameMap GameMap, planetName string) *Planet {
	for _, system := range gameMap.StarSystems {
		for _, planet := range system.Planets {
			if planet.Name == planetName {
				return &planet // Return full planet struct
			}
		}
	}
	return nil // Return nil if not found
}

// GetDistance calculates the distance from the ship to a certain planet by name
func GetDistance(gameMap GameMap, ship Ship, planetName string) int {
	// Get ships coords
	shipCoords := ship.Location.Coordinates

	// Find the planet using helper function
	planet := FindPlanet(gameMap, planetName)
	if planet == nil {
		return -1
	}

	// (X + X, Y + Y, Z + Z) * 10
	distance := (int(math.Abs(float64(shipCoords.X+planet.Coordinates.X))) +
		int(math.Abs(float64(shipCoords.Y+planet.Coordinates.Y))) +
		int(math.Abs(float64(shipCoords.Z+planet.Coordinates.Z)))) * 10

	return distance
}

// returns a random ID string with the given prefix
func generateRandomID(prefix string) string {
	return prefix + strconv.Itoa(rand.Intn(1000000))
}

// ------------------------------
// Full Game Save File Operations
// ------------------------------

// creates a new full game save file using the provided parameters
func CreateNewFullGameSave(difficulty, shipName, startingLocation string) error {
	now := time.Now()

	// Define default missions.
	defaultMissions := Missions{
		Main: []Mission{
			{
				Step:              0,
				Title:             "Rescue Mission",
				Description:       "Rescue the stranded astronaut on a rogue asteroid",
				Status:            "Not Started",
				Location:          "Planet A",
				Income:            1000,
				Requirements:      "None",
				Received:          "Game",
				Category:          "Main",
				TravelTime:        5,
				FuelNeeded:        10,
				DestinationPlanet: "Planet A",
				Dialogue: []string{
					"Commander, we have received a distress signal from a stranded astronaut on a rogue asteroid.",
					"Your mission is to rescue the astronaut and bring them back to safety.",
					"Time is of the essence, Commander. We need you to act quickly.",
				},
			},
		},
		Received: []ReceivedMissionGroup{
			{
				Location: "Mars",
				NPCs: []NPC{
					{
						Name: "Commander Vega",
						Missions: []Mission{
							{
								Title:             "Solar Flare Response",
								Description:       "Monitor and respond to unpredictable solar flare activities.",
								Status:            "In Progress",
								Income:            4000,
								Requirements:      "Shielded Satellite",
								Received:          "Commander Vega",
								Category:          "Received",
								TravelTime:        3,
								FuelNeeded:        30,
								DestinationPlanet: "Mars",
								Dialogue: []string{
									"Commander, a massive solar flare is imminent.",
									"Prepare your shields and adjust your course to minimize damage.",
									"Your swift action is needed to protect our assets.",
								},
							},
						},
					},
				},
			},
		},
	}
	// Define default planets
	defaultGameMap := GameMap{
		StarSystems: []StarSystem{
			{
				SystemID:    "SYS_499172",
				Name:        "Sol",
				Coordinates: Coordinates{X: 0, Y: 0, Z: 0},
				Planets: []Planet{
					{
						PlanetID:    "Earth",
						Name:        "Earth",
						Type:        "Terrestrial",
						Coordinates: Coordinates{X: 0, Y: 0, Z: 0},
					},
					{
						PlanetID:    "Mars",
						Name:        "Mars",
						Type:        "Terrestrial",
						Coordinates: Coordinates{X: 5, Y: 3, Z: 1},
					},
				},
			},
		},
	}

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
			EngineHealth:      100,
			FTLDriveHealth:    70,
			FTLDriveCharge:    0,
			Food:              100,
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
		Missions: defaultMissions,
		GameMap:  defaultGameMap,
	}

	// Wrap the save data in an array (slice) as per your JSON structure.
	saveData := []FullGameSave{fullSave}

	dataBytes, err := json.MarshalIndent(saveData, "", "  ")
	if err != nil {
		return err
	}

	// Ensure the directory exists.
	dir := filepath.Dir(SaveFilePath)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}

	return ioutil.WriteFile(SaveFilePath, dataBytes, 0644)
}

// returns a default FullGameSave structure with initial "new game" values
func DefaultFullGameSave() *FullGameSave {
	now := time.Now()

	defaultMissions := Missions{
		Main: []Mission{
			{
				Step:              0,
				MissionId:         generateRandomID("MISSION_"),
				Title:             "Rescue Mission",
				Description:       "Rescue the stranded astronaut on a rogue asteroid",
				Status:            "Not Started",
				Location:          "Planet A",
				Income:            1000,
				Requirements:      "None",
				Received:          "Game",
				Category:          "Main",
				TravelTime:        5,
				FuelNeeded:        10,
				DestinationPlanet: "Planet A",
				Dialogue: []string{
					"Commander, we have received a distress signal from a stranded astronaut on a rogue asteroid.",
					"Your mission is to rescue the astronaut and bring them back to safety.",
					"Time is of the essence, Commander. We need you to act quickly.",
				},
			},
		},
		Received: []ReceivedMissionGroup{
			{
				Location: "Mars",
				NPCs: []NPC{
					{
						Name: "Commander Vega",
						Missions: []Mission{
							{
								MissionId:         generateRandomID("MISSION_"),
								Title:             "Solar Flare Response",
								Description:       "Monitor and respond to unpredictable solar flare activities.",
								Status:            "In Progress",
								Income:            4000,
								Requirements:      "Shielded Satellite",
								Received:          "Commander Vega",
								Category:          "Received",
								TravelTime:        3,
								FuelNeeded:        30,
								DestinationPlanet: "Mars",
								Dialogue: []string{
									"Commander, a massive solar flare is imminent.",
									"Prepare your shields and adjust your course to minimize damage.",
									"Your swift action is needed to protect our assets.",
								},
							},
						},
					},
				},
			},
		},
	}
	defaultGameMap := GameMap{
		StarSystems: []StarSystem{
			{
				SystemID:    "SYS_499172",
				Name:        "Sol",
				Coordinates: Coordinates{X: 0, Y: 0, Z: 0},
				Planets: []Planet{
					{
						PlanetID:    "Earth",
						Name:        "Earth",
						Type:        "Terrestrial",
						Coordinates: Coordinates{X: 0, Y: 0, Z: 0},
					},
					{
						PlanetID:    "Mars",
						Name:        "Mars",
						Type:        "Terrestrial",
						Coordinates: Coordinates{X: 5, Y: 3, Z: 1},
					},
				},
			},
		},
	}

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
			EngineHealth:      100,
			FTLDriveHealth:    70,
			FTLDriveCharge:    0,
			Food:              100,
			Location: Location{
				StarSystemId: "SYS_0001",
				PlanetId:     "Earth",
				Coordinates:  Coordinates{X: 0, Y: 0, Z: 0},
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
		Missions: defaultMissions,
		GameMap:  defaultGameMap,
	}
}

// checks whether a save file already exists
func SaveExists() bool {
	_, err := os.Stat(SaveFilePath)
	return err == nil
}

// reads the JSON save file and returns the full game data
func LoadFullGameSave() (*FullGameSave, error) {
	dataBytes, err := ioutil.ReadFile(SaveFilePath)
	if err != nil {
		return nil, err
	}
	var saves []FullGameSave
	if err := json.Unmarshal(dataBytes, &saves); err != nil {
		return nil, err
	}
	if len(saves) == 0 {
		return nil, nil
	}
	return &saves[0], nil
}

// writes the current full game save to disk
func SaveGame(save *FullGameSave) error {
	// Wrap the save data in a slice as per your JSON structure.
	saveData := []FullGameSave{*save}

	dataBytes, err := json.MarshalIndent(saveData, "", "  ")
	if err != nil {
		return err
	}

	// Write to a temporary file first.
	tmpFilePath := SaveFilePath + ".tmp"
	if err := ioutil.WriteFile(tmpFilePath, dataBytes, 0644); err != nil {
		return err
	}

	// Rename temporary file to actual save file.
	return os.Rename(tmpFilePath, SaveFilePath)
}
