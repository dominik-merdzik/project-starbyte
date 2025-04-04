package data

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

const SaveFilePath = "GameData/save/save.json"

// We have to manually bump this for each release. We should probably automate this.
const version = "1.0.1-beta"

// ---------------------
// Save File Structures
// ---------------------

type FullGameSave struct {
	GameTitle    string       `json:"gameTitle"`
	GameMetadata GameMetadata `json:"gameMetadata"`
	Player       Player       `json:"player"`
	Ship         Ship         `json:"ship"`
	Crew         []CrewMember `json:"crew"`
	Missions     []Mission    `json:"missions"`
	GameMap      GameMap      `json:"gameMap"`
	Collection   Collection   `json:"collection"`
}

type GameMetadata struct {
	Version            string             `json:"version"`
	DateCreated        string             `json:"dateCreated"`
	LastSaveTime       string             `json:"lastSaveTime"`
	TotalPlayTime      TotalPlayTime      `json:"totalPlayTime"`
	DifficultySettings DifficultySettings `json:"difficultySettings"`
	GameOver           bool               `json:"gameOver"`
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

// ---------------------
// Ship structures
// ---------------------

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
	MaxEngineHealth   int      `json:"maxEngineHealth"`
	HasFTLDrive       bool     `json:"hasFTLDrive"`
	FTLDriveHealth    int      `json:"ftlDriveHealth"`
	FTLDriveCharge    int      `json:"ftlDriveCharge"`
	Food              int      `json:"food"`
	Location          Location `json:"location"`
	Cargo             Cargo    `json:"cargo"`
	Modules           []Module `json:"modules"`
	Upgrades          Upgrades `json:"upgrades"`
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

// ---------------------
// Crew structures
// ---------------------

type CrewRole string

const (
	CrewRolePilot                 CrewRole = "Pilot"
	CrewRoleEngineer              CrewRole = "Engineer"
	CrewRoleScientist             CrewRole = "Scientist"
	CrewRoleMedic                 CrewRole = "Medic"
	CrewRoleSecurityOfficer       CrewRole = "Security Officer"
	CrewRoleNavigator             CrewRole = "Navigator"
	CrewRoleCommunicationsOfficer CrewRole = "Communications Officer"
	CrewRoleMechanic              CrewRole = "Mechanic"
	CrewRoleWeaponsSpecialist     CrewRole = "Weapons Specialist"
	CrewRoleResearchSpecialist    CrewRole = "Research Specialist"
)

// Updated CrewMember: Removed Skills and added Buffs and Debuffs.
type CrewMember struct {
	CrewId          string   `json:"crewId"`
	Name            string   `json:"name"`
	Role            CrewRole `json:"role"`
	Degree          int      `json:"degree"`
	Experience      int      `json:"experience"`
	Morale          int      `json:"morale"`
	Health          int      `json:"health"`
	MasterWorkLevel int      `json:"masterWorkLevel"`
	Buffs           []string `json:"buffs"`
	Debuffs         []string `json:"debuffs"`
	AssignedTaskId  *string  `json:"assignedTaskId"`
}

// ---------------------
// Map structures
// ---------------------

type Location struct {
	StarSystemName string      `json:"starSystemName"`
	PlanetName     string      `json:"planetName"`
	Coordinates    Coordinates `json:"coordinates"`
}

type GameMap struct {
	StarSystems []StarSystem `json:"starSystems"`
}

type StarSystem struct {
	Name    string   `json:"name"`
	Planets []Planet `json:"planets"`
}

type Planet struct {
	Name         string            `json:"name"`
	Type         string            `json:"type"`
	Resources    []Resource        `json:"resources"`
	Coordinates  Coordinates       `json:"coordinates"`
	Requirements []CrewRequirement `json:"requirements"`
}

type Resource struct {
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
}

type CrewRequirement struct {
	Role   string `json:"role"`
	Degree int    `json:"degree"`
	Count  int    `json:"count"`
}

// ---------------------
// Mission structures
// ---------------------

type Mission struct {
	Step         int           `json:"Step,omitempty"`
	Id           int           `json:"Id"`
	Title        string        `json:"Title"`
	Description  string        `json:"Description"`
	Status       MissionStatus `json:"Status"`
	Location     Location      `json:"Location"`
	Income       int           `json:"Income"`
	Requirements string        `json:"Requirements"`
	Received     string        `json:"Received"`
	Category     string        `json:"Category"`
	Dialogue     []string      `json:"dialogue"`
}

type MissionStatus int

const (
	MissionStatusNotStarted MissionStatus = iota
	MissionStatusInProgress
	MissionStatusCompleted
	MissionStatusFailed
	MissionStatusAbandoned
)

func (ms MissionStatus) String() string {
	return [...]string{"Not Started", "In Progress", "Completed", "Failed", "Abandoned"}[ms]
}

// ---------------------
// Collection structures
// ---------------------

type Collection struct {
	MaxCapacity   int                `json:"maxCapacity"`
	UsedCapacity  int                `json:"usedCapacity"`
	Items         []CollectionItem   `json:"items"`
	ResearchNotes []ResearchNoteTier `json:"researchNotes"`
}

type CollectionItem struct {
	ItemId      string `json:"itemId"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Quantity    int    `json:"quantity"`
	Tier        int    `json:"tier,omitempty"`
}

type ResearchNoteTier struct {
	Name     string `json:"name"`
	Blurb    string `json:"blurb"`
	Tier     int    `json:"tier"`
	XP       int    `json:"xp"`
	Quantity int    `json:"quantity"`
}

// -------------------
// Buffs & Debuffs Pools and Modifier Logic
// -------------------

// BuffPool contains the possible buffs a crew member can receive
var BuffPool = []string{
	"Sharp Shooter",
	"Quick Reflexes",
	"Enhanced Strength",
	"Iron Will",
	"Expert Navigator",
}

// DebuffPool contains the possible debuffs a crew member can receive
var DebuffPool = []string{
	"Sluggish",
	"Tired",
	"Unfocused",
	"Injured",
	"Distracted",
}

// AwardModifier awards a buff or debuff every time the crew member crosses a 10-level threshold
// For each threshold passed, there is a 60% chance for a buff and a 40% chance for a debuff
// It returns a receipt message summarizing the awarded modifiers
func AwardModifier(crew *CrewMember, oldDegree, newDegree int) string {
	receipt := ""
	oldThreshold := oldDegree / 10
	newThreshold := newDegree / 10
	for i := oldThreshold + 1; i <= newThreshold; i++ {
		roll := rand.Intn(100)
		if roll < 60 {
			buff := BuffPool[rand.Intn(len(BuffPool))]
			crew.Buffs = append(crew.Buffs, buff)
			receipt += fmt.Sprintf("Received buff: '%s'\n", buff)
		} else {
			debuff := DebuffPool[rand.Intn(len(DebuffPool))]
			crew.Debuffs = append(crew.Debuffs, debuff)
			receipt += fmt.Sprintf("Received debuff: '%s'\n", debuff)
		}
	}
	return receipt
}

// ---------------------
// Helper Functions
// ---------------------

func generateRandomID(prefix string) string {
	return prefix + strconv.Itoa(rand.Intn(1000000))
}

func DefaultCollection() Collection {
	return Collection{
		MaxCapacity:  100,
		UsedCapacity: 0,
		Items: []CollectionItem{
			{
				ItemId:      generateRandomID("ITEM_"),
				Name:        "Space Debris",
				Description: "Miscellaneous space junk.",
				Quantity:    1,
				Tier:        1,
			},
			{
				ItemId:      generateRandomID("ITEM_"),
				Name:        "Exotic Matter Sample",
				Description: "A sample of unknown exotic matter.",
				Quantity:    1,
				Tier:        2,
			},
			{
				ItemId:      generateRandomID("ITEM_"),
				Name:        "Strange Artifact",
				Description: "An unidentifiable artifact of unknown origin.",
				Quantity:    1,
				Tier:        5,
			},
		},
		ResearchNotes: []ResearchNoteTier{
			{
				Name:     "Rough Scribbles",
				Blurb:    "These are your earliest musings—quick sketches and fragmented ideas jotted down in the heat of discovery.",
				Tier:     1,
				XP:       100,
				Quantity: 0,
			},
			{
				Name:     "Field Observations",
				Blurb:    "Compiled during your initial forays into uncharted territory, these notes capture raw experiences that hint at a larger mystery.",
				Tier:     2,
				XP:       200,
				Quantity: 0,
			},
			{
				Name:     "Experimental Logs",
				Blurb:    "With a bit more structure, these offer a clearer look at the phenomena you're unraveling.",
				Tier:     3,
				XP:       300,
				Quantity: 0,
			},
			{
				Name:     "Analytical Reports",
				Blurb:    "Now your notes take on a more refined methodical filled with insightful analysis that bridges observation with theory.",
				Tier:     4,
				XP:       400,
				Quantity: 0,
			},
			{
				Name:     "Breakthrough Manuscripts",
				Blurb:    "The pinnacle of your research journey, these combine rigorous data and innovative thought to reveal groundbreaking insights that could change everything.",
				Tier:     5,
				XP:       500,
				Quantity: 0,
			},
		},
	}
}

// -------------------------------------------
// creates a new game file with default values
// -------------------------------------------

func CreateNewFullGameSave(difficulty, shipName, playerName string) error {
	now := time.Now()

	defaultMissions := []Mission{
		{
			Step:         0,
			Id:           0,
			Title:        "Rescue Mission",
			Description:  "Rescue the stranded astronaut on a rogue asteroid",
			Status:       MissionStatusNotStarted,
			Location:     Location{StarSystemName: "Sol", PlanetName: "Asteroid X", Coordinates: Coordinates{X: 0, Y: 0, Z: 1}},
			Income:       1000,
			Requirements: "Pilot",
			Received:     "Commander Vega (ISS)",
			Category:     "Main",
			Dialogue: []string{
				"Commander, we've received a distress signal from the outer asteroid belt. Faint, but repeating.",
				"There's a lone astronaut stranded out there on what looks like a rogue asteroid, designation AX-7.",
				"Reports suggest the asteroid's trajectory is... unnatural. It doesn't match any known orbital mechanics.",
				"The astronaut might have seen something out there... something important.",
				"Get them home safely. Bring them back to the ISS [0,0,0] for debriefing. This could be bigger than just a rescue.",
			},
		},
		{
			Id:           11,
			Title:        "Solar Flare Response",
			Description:  "Monitor and respond to unpredictable solar flare activities.",
			Status:       MissionStatusNotStarted,
			Location:     Location{StarSystemName: "Sol", PlanetName: "Mars", Coordinates: Coordinates{X: 5, Y: 3, Z: 1}},
			Income:       4000,
			Requirements: "Shielded Satellite",
			Received:     "Commander Vega",
			Category:     "Received",
			Dialogue: []string{
				"Commander, a massive solar flare is imminent.",
				"Prepare your shields and adjust your course to minimize damage.",
				"Your swift action is needed to protect our assets.",
			},
		},
	}

	defaultGameMap := GameMap{
		StarSystems: []StarSystem{
			{
				Name: "Sol",
				Planets: []Planet{
					{
						Name:        "ISS",
						Type:        "Space Station",
						Coordinates: Coordinates{X: 0, Y: 0, Z: 0},
						Requirements: []CrewRequirement{
							{Role: "Pilot", Degree: 1, Count: 1},
						},
					},
					{
						Name:        "Earth",
						Type:        "Terrestrial",
						Coordinates: Coordinates{X: 2, Y: 4, Z: 5},
						Requirements: []CrewRequirement{
							{Role: "Pilot", Degree: 1, Count: 1},
							{Role: "Engineer", Degree: 1, Count: 1},
						},
					},
					{
						Name:        "Mars",
						Type:        "Terrestrial",
						Coordinates: Coordinates{X: -3, Y: -4, Z: -3},
						Requirements: []CrewRequirement{
							{Role: "Engineer", Degree: 1, Count: 1},
						},
					},
					{
						Name:        "Jupiter",
						Type:        "Gas Giant",
						Coordinates: Coordinates{X: 9, Y: -20, Z: 5},
						Requirements: []CrewRequirement{
							{Role: "Engineer", Degree: 1, Count: 1},
						},
					},
					{
						Name:        "Saturn",
						Type:        "Gas Giant",
						Coordinates: Coordinates{X: 20, Y: 30, Z: 10},
						Requirements: []CrewRequirement{
							{Role: "Engineer", Degree: 1, Count: 1},
						},
					},
				},
			},
			{
				Name: "Alpha Centauri",
				Planets: []Planet{
					{
						Name:        "Proxima b",
						Type:        "Terrestrial",
						Coordinates: Coordinates{X: 1, Y: 2, Z: 3},
						Requirements: []CrewRequirement{
							{Role: "Pilot", Degree: 1, Count: 1},
							{Role: "Engineer", Degree: 2, Count: 1},
						},
					},
					{
						Name:        "Alpha Centauri Bb",
						Type:        "Gas Giant",
						Coordinates: Coordinates{X: 2, Y: 1, Z: 0},
						Requirements: []CrewRequirement{
							{Role: "Pilot", Degree: 1, Count: 1},
							{Role: "Engineer", Degree: 2, Count: 2},
						},
					},
				},
			},
			{
				Name: "Sirius",
				Planets: []Planet{
					{
						Name:        "Sirius I",
						Type:        "Terrestrial",
						Coordinates: Coordinates{X: -1, Y: 0, Z: 2},
						Requirements: []CrewRequirement{
							{Role: "Pilot", Degree: 2, Count: 1},
						},
					},
					{
						Name:        "Sirius II",
						Type:        "Gas Giant",
						Coordinates: Coordinates{X: -2, Y: 3, Z: 1},
						Requirements: []CrewRequirement{
							{Role: "Engineer", Degree: 2, Count: 1},
						},
					},
					{
						Name:        "Sirius III",
						Type:        "Ice Giant",
						Coordinates: Coordinates{X: -3, Y: 3, Z: 3},
						Requirements: []CrewRequirement{
							{Role: "Scientist", Degree: 3, Count: 1},
						},
					},
				},
			},
			{
				Name: "Vega",
				Planets: []Planet{
					{
						Name:        "Vega I",
						Type:        "Terrestrial",
						Coordinates: Coordinates{X: 0, Y: 1, Z: -1},
						Requirements: []CrewRequirement{
							{Role: "Pilot", Degree: 1, Count: 1},
							{Role: "Engineer", Degree: 1, Count: 1},
						},
					},
					{
						Name:        "Vega II",
						Type:        "Gas Giant",
						Coordinates: Coordinates{X: 1, Y: 1, Z: 1},
						Requirements: []CrewRequirement{
							{Role: "Engineer", Degree: 2, Count: 2},
						},
					},
				},
			},
		},
	}

	fullSave := FullGameSave{
		GameTitle: "Project Starbyte",
		GameMetadata: GameMetadata{
			Version:      version,
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
			PlayerName:       playerName,
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
			MaxFuel:           100,
			EngineHealth:      100,
			MaxEngineHealth:   100,
			HasFTLDrive:       false,
			FTLDriveHealth:    10,
			FTLDriveCharge:    0,
			Food:              100,
			Location: Location{
				StarSystemName: "Sol",
				PlanetName:     "ISS",
				Coordinates:    Coordinates{X: 0, Y: 0, Z: 0},
			},
			Cargo: Cargo{
				Capacity:     100,
				UsedCapacity: 2,
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
					MaxLevel:     10,
				},
				WeaponSystems: UpgradeLevel{
					CurrentLevel: 0,
					MaxLevel:     10,
				},
				CargoExpansion: UpgradeLevel{
					CurrentLevel: 0,
					MaxLevel:     10,
				},
			},
		},
		Crew: []CrewMember{
			{
				CrewId:          generateRandomID("CREW_"),
				Name:            "Alice",
				Role:            CrewRolePilot,
				Degree:          1,
				Experience:      0,
				Morale:          100,
				Health:          100,
				MasterWorkLevel: 0,
				Buffs:           []string{},
				Debuffs:         []string{},
				AssignedTaskId:  nil,
			},
			{
				CrewId:          generateRandomID("CREW_"),
				Name:            "Bob",
				Role:            CrewRoleEngineer,
				Degree:          1,
				Experience:      0,
				Morale:          100,
				Health:          100,
				MasterWorkLevel: 0,
				Buffs:           []string{},
				Debuffs:         []string{},
				AssignedTaskId:  nil,
			},
		},
		Missions:   defaultMissions,
		GameMap:    defaultGameMap,
		Collection: DefaultCollection(),
	}

	saveData := []FullGameSave{fullSave}

	dataBytes, err := json.MarshalIndent(saveData, "", "  ")
	if err != nil {
		return err
	}

	dir := filepath.Dir(SaveFilePath)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}

	return ioutil.WriteFile(SaveFilePath, dataBytes, 0644)
}

// ---------------------
// Save File Operations
// ---------------------

func SaveExists() bool {
	_, err := os.Stat(SaveFilePath)
	return err == nil
}

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

func SaveGame(save *FullGameSave) error {
	saveData := []FullGameSave{*save}

	dataBytes, err := json.MarshalIndent(saveData, "", "  ")
	if err != nil {
		return err
	}

	tmpFilePath := SaveFilePath + ".tmp"
	if err := ioutil.WriteFile(tmpFilePath, dataBytes, 0644); err != nil {
		return err
	}

	return os.Rename(tmpFilePath, SaveFilePath)
}

func CheckCrewRequirement(crewList []CrewMember, req CrewRequirement) bool {
	qualifiedCount := 0
	for _, crewMember := range crewList {
		// Check if the crew member's role matches the requirement's role
		// and if their degree meets or exceeds the required degree.
		// Assumes req.Role is string and crewMember.Role is data.CrewRole type
		if string(crewMember.Role) == req.Role && crewMember.Degree >= req.Degree {
			qualifiedCount++
		}
	}
	// Requirement is met if we found at least the required number of qualified crew.
	return qualifiedCount >= req.Count
}
