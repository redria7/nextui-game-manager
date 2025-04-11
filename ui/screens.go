package ui

import (
	"bytes"
	"errors"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/ui"
	"go.uber.org/zap"
	"nextui-game-manager/models"
	"nextui-game-manager/state"
	"os"
	"os/exec"
	"path/filepath"
	"qlova.tech/sum"
	"strings"
)

var Screens = sum.Int[models.Screen]{}.Sum()

var ScreenFuncs = map[sum.Int[models.Screen]]func() shared.Selection{
	Screens.MainMenu:    mainMenuScreen,
	Screens.Loading:     loading,
	Screens.GamesList:   gamesList,
	Screens.SearchBox:   searchBox,
	Screens.Actions:     actionScreen,
	Screens.Confirm:     confirmationScreen,
	Screens.DownloadArt: downloadArtScreen,
}

func SetScreen(screen sum.Int[models.Screen]) {
	tempAppState := state.GetAppState()
	tempAppState.CurrentScreen = screen
	state.UpdateAppState(tempAppState)
}

func mainMenuScreen() shared.Selection {
	appState := state.GetAppState()

	menu := ""

	var romDirectories []string
	for _, dir := range appState.RomDirectories {
		romDirectories = append(romDirectories, dir.DisplayName)
	}

	menu = strings.Join(romDirectories, "\n")

	var extraArgs []string
	extraArgs = append(extraArgs, "--cancel-text", "QUIT")

	return ui.DisplayMinUiList(menu, "text", "Game Manager", extraArgs...)
}

func loading() shared.Selection {
	logger := common.GetLoggerInstance()
	appState := state.GetAppState()

	logger.Debug("Selected ROM Directory", zap.String("ROM Directory", appState.CurrentSection.LocalDirectory))

	romEntries, err := os.ReadDir(appState.CurrentSection.LocalDirectory)
	if err != nil {
		return shared.Selection{Code: 1}
	}

	var roms shared.Items

	for _, entry := range romEntries {
		if !entry.IsDir() {
			roms = append(roms, shared.Item{
				Filename: entry.Name(),
			})
		}
	}

	for _, rom := range roms {
		logger.Debug("Rom", zap.String("com", rom.Filename))
	}

	appState.CurrentItemsList = roms

	state.UpdateAppState(appState)

	return shared.Selection{Code: 0}
}

func gamesList() shared.Selection {
	appState := state.GetAppState()

	title := appState.CurrentSection.Name
	itemList := appState.CurrentItemsList

	var extraArgs []string
	extraArgs = append(extraArgs, "--confirm-text", "SELECT")

	if appState.SearchFilter != "" {
		title = "[Search: \"" + appState.SearchFilter + "\"]"
		extraArgs = append(extraArgs, "--cancel-text", "CLEAR SEARCH")
		itemList = filterList(itemList, appState.SearchFilter)
	}

	if len(itemList) == 0 {
		return shared.Selection{Code: 404}
	}

	var itemEntries []string
	for _, item := range itemList {
		itemName := strings.TrimSuffix(item.Filename, filepath.Ext(item.Filename))
		itemEntries = append(itemEntries, itemName)
	}

	if len(itemEntries) > 500 {
		itemEntries = itemEntries[:500]
	}

	return ui.DisplayMinUiListWithAction(strings.Join(itemEntries, "\n"), "text", title, "SEARCH", extraArgs...)
}

func searchBox() shared.Selection {
	logger := common.GetLoggerInstance()

	args := []string{"--title", "Game Search"}

	cmd := exec.Command("minui-keyboard", args...)
	cmd.Env = os.Environ()
	cmd.Env = os.Environ()

	var stdoutbuf, stderrbuf bytes.Buffer
	cmd.Stdout = &stdoutbuf
	cmd.Stderr = &stderrbuf

	if errors.Is(cmd.Err, exec.ErrDot) {
		cmd.Err = nil
	}

	err := cmd.Start()
	if err != nil {
		logger.Fatal("failed to start minui-keyboard", zap.Error(err))
	}

	err = cmd.Wait()
	if err != nil && cmd.ProcessState.ExitCode() == 1 {
		logger.Error("Error with keyboard", zap.String("error", stderrbuf.String()))
		ShowMessage("Unable to open keyboard!", "3")
		return shared.Selection{Code: 1}
	}

	outValue := stdoutbuf.String()
	_ = stderrbuf.String()

	return shared.Selection{Value: strings.TrimSpace(outValue), Code: cmd.ProcessState.ExitCode()}
}

func actionScreen() shared.Selection {
	appState := state.GetAppState()

	return ui.DisplayMinUiList(strings.Join(models.ActionKeys, "\n"), "text", appState.SelectedFile)
}

func confirmationScreen() shared.Selection {
	return shared.Selection{}
}

func downloadArtScreen() shared.Selection {
	return shared.Selection{}
}
