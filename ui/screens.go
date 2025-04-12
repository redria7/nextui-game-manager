package ui

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/ui"
	"go.uber.org/zap"
	"nextui-game-manager/models"
	"nextui-game-manager/state"
	"nextui-game-manager/utils"
	"os"
	"os/exec"
	"path/filepath"
	"qlova.tech/sum"
	"strings"
	"time"
)

var Screens = sum.Int[models.Screen]{}.Sum()

var ScreenFuncs = map[sum.Int[models.Screen]]func() shared.Selection{
	Screens.MainMenu:    mainMenuScreen,
	Screens.Loading:     loading,
	Screens.GamesList:   gamesList,
	Screens.SearchBox:   searchBox,
	Screens.Actions:     actionScreen,
	Screens.Confirm:     confirmationScreen,
	Screens.RenameRom:   renameRomScreen,
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
		romDirectories = append(romDirectories, strings.TrimSpace(dir.DisplayName))
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

	err := utils.RefreshRomsList()
	if err != nil {
		return shared.Selection{Code: 1}
	}

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
		itemList = utils.FilterList(itemList, appState.SearchFilter)
	}

	if len(itemList) == 0 {
		return shared.Selection{Code: 404}
	}

	var itemEntries []string

	for _, item := range itemList {
		itemName := strings.TrimSuffix(item.Filename, filepath.Ext(item.Filename))
		itemEntries = append(itemEntries, itemName)
	}

	state.UpdateAppState(appState)

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
	logger := common.GetLoggerInstance()
	appState := state.GetAppState()

	existingArtFilename, err := utils.FindExistingArt()
	if err != nil {
		logger.Error("failed to find existing arts", zap.Error(err))
	}

	hasGameTrackerData := utils.HasGameTrackerData()

	actions := models.ActionKeys

	if existingArtFilename == "" {
		actions = utils.InsertIntoSlice(actions, 1, "Download Art")
	} else {
		actions = utils.InsertIntoSlice(actions, 1, "Delete Art")
	}

	if hasGameTrackerData {
		actions = utils.InsertIntoSlice(actions, 2, "Clear Game Tracker")
	}

	return ui.DisplayMinUiList(strings.Join(actions, "\n"), "text", appState.SelectedFile)
}

func confirmationScreen() shared.Selection {
	appState := state.GetAppState()

	actionMessage := models.ActionMessages[appState.SelectedAction]

	message := fmt.Sprintf("%s %s?", actionMessage, strings.Split(appState.SelectedFile, ".")[0])

	code, err := ui.ShowMessageWithOptions(message, "0",
		"--confirm-text", "DO IT!",
		"--confirm-show", "true",
		"--confirm-button", "X",
		"--cancel-show", "true",
		"--cancel-text", "CHANGED MY MIND",
	)

	if err != nil {
		logger := common.GetLoggerInstance()
		logger.Info("Oh no", zap.Error(err))
	}

	return shared.Selection{Code: code}
}

func renameRomScreen() shared.Selection {
	logger := common.GetLoggerInstance()
	appState := state.GetAppState()

	args := []string{"--initial-value", appState.SelectedFile, "--title", "Rename ROM", "--show-hardware-group"}

	logger.Info("Opening Rename Keyboard", zap.Strings("args", args))

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
		logger.Fatal("failed to start minui-keyboard", zap.Error(err),
			zap.String("stdout", stdoutbuf.String()), zap.String("error", stderrbuf.String()))
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

func downloadArtScreen() shared.Selection {
	logger := common.GetLoggerInstance()

	ctx := context.Background()
	ctxWithCancel, cancel := context.WithCancel(ctx)
	defer cancel()

	args := []string{"--message", "Attempting to download art...", "--timeout", "-1"}
	cmd := exec.CommandContext(ctxWithCancel, "minui-presenter", args...)

	err := cmd.Start()
	if err != nil && cmd.ProcessState.ExitCode() != -1 {
		logger.Fatal("Error with starting miniui-presenter download message", zap.Error(err))
	}

	time.Sleep(1000 * time.Millisecond)

	exitCode := 0

	go func() {
		res := utils.FindArt()
		if !res {
			logger.Error("Could not find art!", zap.Error(err))
			exitCode = 1
		}

		cancel()
	}()

	err = cmd.Wait()
	if err != nil && cmd.ProcessState.ExitCode() != -1 {
		logger.Fatal("Error with minui-presenter display of download message", zap.Error(err))
	}

	return shared.Selection{Code: exitCode}
}
