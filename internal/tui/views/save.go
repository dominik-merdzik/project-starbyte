package views

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sync"
	"time"
)

// SaveFile represents a structure to hold data and manage concurrent access
type SaveFile struct {
	Data  map[string]interface{} `json:"data"`
	mutex sync.Mutex
}

// Global instance of SaveFile to store data
var globalSaveFile *SaveFile = NewSaveFile()

// Queue to store field additions before processing
var saveQueue map[string]interface{} = make(map[string]interface{})
var queueMutex sync.Mutex

// Save filename
const saveFilename = "savefile.json"

// NewSaveFile initializes a new SaveFile instance
func NewSaveFile() *SaveFile {
	return &SaveFile{Data: make(map[string]interface{})}
}

// AddField adds a new field to the SaveFile in a thread-safe manner
func (s *SaveFile) AddField(fieldName string, value interface{}) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.Data[fieldName] = value
}

// SaveToFile saves the SaveFile data to a JSON file
func (s *SaveFile) SaveToFile(filename string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	file, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, file, 0644)
}

// LoadFromFile loads data from a JSON file into the SaveFile
func (s *SaveFile) LoadFromFile(filename string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	return json.Unmarshal(file, s)
}

// QueueField adds a field update to the queue, preventing duplicates
func QueueField(fieldName string, valuePtr interface{}) {
	queueMutex.Lock()
	defer queueMutex.Unlock()

	saveQueue[fieldName] = valuePtr
}

// ProcessQueue applies all queued field additions to the SaveFile
func ProcessQueue() {
	queueMutex.Lock()
	defer queueMutex.Unlock()

	for fieldName, value := range saveQueue {
		globalSaveFile.AddField(fieldName, value)
	}
	// Clear the queue after processing
	saveQueue = make(map[string]interface{})
}

// Save processes all queued field additions and writes to the file
func SaveGame() error {
	ProcessQueue()
	return globalSaveFile.SaveToFile(saveFilename)
}

// LoadGame loads the save file and reconstructs structured game data
func LoadGame() error {
	return globalSaveFile.LoadFromFile(saveFilename)
}

// LoadGameState retrieves the saved game state and unmarshals it into a struct
func LoadGameState(gameState interface{}) error {
	if err := LoadGame(); err != nil {
		return err
	}

	if data, exists := globalSaveFile.Data["game_state"]; exists {
		jsonData, _ := json.Marshal(data)
		return json.Unmarshal(jsonData, gameState)
	}
	return fmt.Errorf("game state not found in save file")
}

// GetSavedField retrieves a specific field from the save file and unmarshals it into target
func GetSavedField(fieldName string, target interface{}) error {
	globalSaveFile.mutex.Lock()
	defer globalSaveFile.mutex.Unlock()

	if data, exists := globalSaveFile.Data[fieldName]; exists {
		jsonData, _ := json.Marshal(data)
		return json.Unmarshal(jsonData, target)
	}
	return fmt.Errorf("field %s not found in save file", fieldName)
}

// RetrieveField retrieves a field and returns it as an interface{}
func RetrieveField(fieldName string) (interface{}, error) {
	globalSaveFile.mutex.Lock()
	defer globalSaveFile.mutex.Unlock()

	if data, exists := globalSaveFile.Data[fieldName]; exists {
		return data, nil
	}
	return nil, fmt.Errorf("field %s not found in save file", fieldName)
}

// AutoSave triggers a save after a specified duration
func AutoSave(interval time.Duration) {
	for {
		time.Sleep(interval)
		fmt.Println("Auto-saving...")
		if err := SaveGame(); err != nil {
			fmt.Println("Error auto-saving file:", err)
		} else {
			fmt.Println("Auto-save complete!")
		}
	}
}
