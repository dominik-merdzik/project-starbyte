// utilities/state_manager.go
package utilities

import (
	"fmt"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dominik-merdzik/project-starbyte/internal/data"
)

// global mutex to ensure one save at a time
var saveMutex sync.Mutex

// message types for the save result
type SaveSuccessMsg struct{}
type SaveRetryMsg struct{}

// ManualSave (called by user use PushSave() for program driven saving) attempts to save the game in a thread-safe manner
// The syncFunc parameter is a callback to update the game state (e.g., calling g.syncSaveData())
func ManualSave(gameSave *data.FullGameSave, syncFunc func()) tea.Cmd {
	return func() tea.Msg {
		saveMutex.Lock()
		defer saveMutex.Unlock()

		// Update the game state prior to saving
		syncFunc()

		// Attempt to write the game save.
		if err := data.SaveGame(gameSave); err != nil {
			fmt.Println("Error saving game:", err)
			// Return a retry message if there's an error
			return SaveRetryMsg{}
		}
		// Return a success message if saving was successful
		return SaveSuccessMsg{}
	}
}

// PushSave (called by program) immediately attempts to save the provided gameSave data.
func PushSave(gameSave *data.FullGameSave, syncFunc func()) tea.Cmd {
	return func() tea.Msg {
		saveMutex.Lock()
		defer saveMutex.Unlock()

		syncFunc()

		if err := data.SaveGame(gameSave); err != nil {
			fmt.Println("Error pushing save:", err)
			return SaveRetryMsg{}
		}
		return nil // no message to return
	}
}

// RetryManualSave schedules a retry of ManualSave after 2 seconds
func RetryManualSave(gameSave *data.FullGameSave, syncFunc func()) tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return ManualSave(gameSave, syncFunc)()
	})
}
