package state

import (
	"fmt"
	"go.uber.org/atomic"
	"gopkg.in/yaml.v3"
	"nextui-game-manager/models"
	"os"
	"sync"
)

var appState atomic.Pointer[models.AppState]
var onceAppState sync.Once

func LoadConfig() (*models.Config, error) {
	data, err := os.ReadFile("config.yml")
	if err != nil {
		return nil, fmt.Errorf("reading config.yml: %w", err)
	}

	var config models.Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("parsing config.yml: %w", err)
	}

	return &config, nil
}

func GetAppState() *models.AppState {
	onceAppState.Do(func() {
		appState.Store(&models.AppState{})
	})
	return appState.Load()
}

func UpdateAppState(newAppState *models.AppState) {
	appState.Store(newAppState)
}

func SetConfig(config *models.Config) {
	temp := GetAppState()
	temp.Config = config

	UpdateAppState(temp)
}

func AddNewMenuPosition() {
	temp := GetAppState()
	temp.MenuPositionList = append(temp.MenuPositionList, models.MenuPositionPointer{
		SelectedIndex:    0,
		SelectedPosition: 0,
	})
	UpdateAppState(temp)
}

func UpdateCurrentMenuPosition(newIndex int, newPosition int) {
	temp := GetAppState()
	temp.MenuPositionList[len(temp.MenuPositionList)-1] = models.MenuPositionPointer{
		SelectedIndex:    newIndex,
		SelectedPosition: newPosition,
	}
	UpdateAppState(temp)
}

func RemoveMenuPositions(positionCount int) {
	temp := GetAppState()
	listLength := len(temp.MenuPositionList)

	EndPosition := 0
	if positionCount < 0 {
		EndPosition = -1 * positionCount
		if EndPosition > listLength {
			return
		}
	} else {
		if positionCount > listLength {
			positionCount = listLength
		}
		EndPosition = listLength - positionCount
	}

	temp.MenuPositionList = temp.MenuPositionList[:EndPosition]
	UpdateAppState(temp)
}

func ReturnToMain() {
	RemoveMenuPositions(-1)
}

func ReturnToArchiveManagement() {
	RemoveMenuPositions(-3)
}

func ReturnToCollectionManagement() {
	RemoveMenuPositions(-3)
}

func GetCurrentMenuPosition() (int, int) {
	tempList := GetAppState().MenuPositionList
	if len(tempList) <= 0 {
		AddNewMenuPosition()
		tempList = GetAppState().MenuPositionList
	}

	currentPosition := tempList[len(tempList)-1]
	selectedIndex := currentPosition.SelectedIndex
	selectedPosition := currentPosition.SelectedPosition

	selectedPosition = max(0, selectedIndex-selectedPosition)

	return selectedIndex, selectedPosition
}
