package ui

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/filebrowser"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	commonUI "github.com/UncleJunVIP/nextui-pak-shared-functions/ui"
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

var ScreenFuncs = map[sum.Int[models.Screen]]func() (shared.ListSelection, error){
	Screens.MainMenu:             mainMenuScreen,
	Screens.Loading:              loading,
	Screens.GamesList:            gamesList,
	Screens.CollectionsList:      collectionList,
	Screens.CollectionOptions:    collectionOptions,
	Screens.CollectionManagement: collectionManagement,
	Screens.RenameCollection:     renameCollection,
	Screens.SearchBox:            searchBox,
	Screens.Actions:              actionScreen,
	Screens.AddToCollection:      addToCollection,
	Screens.Confirm:              confirmationScreen,
	Screens.RenameRom:            renameRomScreen,
	Screens.DownloadArt:          downloadArtScreen,
}

func SetScreen(screen sum.Int[models.Screen]) {
	tempAppState := state.GetAppState()
	tempAppState.CurrentScreen = screen
	state.UpdateAppState(tempAppState)
}

func mainMenuScreen() (shared.ListSelection, error) {
	appState := state.GetAppState()

	var romDirectories shared.Items
	for _, dir := range appState.RomDirectories {
		romDirectories = append(romDirectories, shared.Item{
			DisplayName: strings.TrimSpace(dir.DisplayName),
		})
	}

	var extraArgs []string
	extraArgs = append(extraArgs, "--cancel-text", "QUIT")

	return commonUI.DisplayList(romDirectories, "Game Manager", "", extraArgs...)
}

func loading() (shared.ListSelection, error) {
	logger := common.GetLoggerInstance()
	appState := state.GetAppState()

	logger.Debug("Selected ROM Directory", zap.String("ROM Directory", appState.CurrentSection.LocalDirectory))

	err := utils.RefreshRomsList()
	if err != nil {
		return shared.ListSelection{ExitCode: 1}, err
	}

	return shared.ListSelection{ExitCode: 0}, err
}

func gamesList() (shared.ListSelection, error) {
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
		return shared.ListSelection{ExitCode: 404}, nil
	}

	var directoryEntries shared.Items
	var itemEntries shared.Items

	for _, item := range itemList {
		if strings.HasPrefix(item.Filename, ".") { // Skip hidden files
			continue
		}

		itemName := strings.TrimSuffix(item.Filename, filepath.Ext(item.Filename))

		if item.IsDirectory {
			itemName = "/" + itemName
			directoryEntries = append(directoryEntries, shared.Item{
				DisplayName: itemName,
			})
			continue
		}

		itemEntries = append(itemEntries, shared.Item{
			DisplayName: itemName,
		})
	}

	allEntries := append(directoryEntries, itemEntries...)

	return commonUI.DisplayList(allEntries, title, "SEARCH", extraArgs...)
}

func collectionList() (shared.ListSelection, error) {
	appState := state.GetAppState()

	title := "Collections"

	fb := filebrowser.NewFileBrowser(common.GetLoggerInstance())
	err := fb.CWD(common.CollectionDirectory)
	if err != nil {
		_, _ = commonUI.ShowMessage("Unable to fetch Collection directories! Quitting!", "3")
		common.LogStandardFatal("Error loading fetching Collection directories", err)
	}

	if len(fb.Items) == 0 {
		return shared.ListSelection{ExitCode: 404}, nil
	}

	itemList := fb.Items
	var collectionDirectory []shared.RomDirectory
	collectionDirectoryMap := make(map[string]shared.RomDirectory)

	for _, item := range fb.Items {
		romDirectory := shared.RomDirectory{
			DisplayName: item.DisplayName,
			Tag:         item.Tag,
			Path:        item.Path,
		}
		collectionDirectory = append(collectionDirectory, romDirectory)
		collectionDirectoryMap[item.DisplayName] = romDirectory
	}

	appState.CollectionDirectories = collectionDirectory
	appState.CollectionDirectoryMap = collectionDirectoryMap

	state.UpdateAppState(appState)

	var extraArgs []string
	extraArgs = append(extraArgs, "--confirm-text", "SELECT")

	if appState.SearchFilter != "" {
		title = "[Search: \"" + appState.SearchFilter + "\"]"
		extraArgs = append(extraArgs, "--cancel-text", "CLEAR SEARCH")
		itemList = utils.FilterList(itemList, appState.SearchFilter)
	}

	if len(itemList) == 0 {
		return shared.ListSelection{ExitCode: 404}, nil
	}

	var itemEntries shared.Items

	for _, item := range itemList {
		itemName := strings.TrimSuffix(item.Filename, filepath.Ext(item.Filename))
		itemEntries = append(itemEntries, shared.Item{
			DisplayName: itemName,
		})
	}

	state.UpdateAppState(appState)

	return commonUI.DisplayList(itemEntries, title, "", extraArgs...)
}

func collectionOptions() (shared.ListSelection, error) {
	title := fmt.Sprintf("Collection Options: %s", state.GetAppState().CurrentSection.Name)

	var extraArgs []string
	extraArgs = append(extraArgs, "--confirm-text", "SELECT")

	var actions shared.Items
	for _, action := range models.CollectionActionKeys {
		actions = append(actions, shared.Item{DisplayName: action})
	}

	return commonUI.DisplayList(actions, title, "", extraArgs...)
}

func renameCollection() (shared.ListSelection, error) {
	logger := common.GetLoggerInstance()
	appState := state.GetAppState()

	args := []string{"--initial-value", appState.CurrentSection.Name, "--title", "Rename ROM", "--show-hardware-group"}

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
		return shared.ListSelection{ExitCode: 1}, err
	}

	outValue := stdoutbuf.String()
	_ = stderrbuf.String()

	return shared.ListSelection{Value: strings.TrimSpace(outValue), ExitCode: cmd.ProcessState.ExitCode()}, nil
}

func collectionManagement() (shared.ListSelection, error) {
	appState := state.GetAppState()

	title := fmt.Sprintf("Collection: %s", appState.CurrentSection.Name)

	collectionList, err := utils.LoadCollectionList(appState.CurrentSection.CollectionFilePath)
	if err != nil {
		return shared.ListSelection{}, err
	}

	var currentItems shared.Items
	collectionItemsMap := make(map[string]string)

	for name, path := range collectionList {
		collectionItem := shared.Item{
			DisplayName: name,
			Path:        path,
		}
		currentItems = append(currentItems, collectionItem)
		collectionItemsMap[collectionItem.DisplayName] = path
	}

	appState.CurrentItemsList = currentItems
	appState.CurrentItemListWithExtensionMap = collectionItemsMap

	state.UpdateAppState(appState)

	var extraArgs []string
	extraArgs = append(extraArgs, "--confirm-text", "SELECT")

	if len(collectionList) == 0 {
		return shared.ListSelection{ExitCode: 404}, nil
	}

	return commonUI.DisplayList(currentItems, title, "OPTIONS", extraArgs...)
}

func searchBox() (shared.ListSelection, error) {
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
		return shared.ListSelection{ExitCode: 1}, err
	}

	outValue := stdoutbuf.String()
	_ = stderrbuf.String()

	return shared.ListSelection{Value: strings.TrimSpace(outValue), ExitCode: cmd.ProcessState.ExitCode()}, nil
}

func actionScreen() (shared.ListSelection, error) {
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

	var actionEntries shared.Items
	for _, action := range actions {
		actionEntries = append(actionEntries, shared.Item{DisplayName: action})
	}

	return commonUI.DisplayList(actionEntries, appState.SelectedFile, "")
}

func addToCollection() (shared.ListSelection, error) {
	return shared.ListSelection{ExitCode: 0}, nil
}

func confirmationScreen() (shared.ListSelection, error) {
	appState := state.GetAppState()

	actionMessage := models.ActionMessages[appState.SelectedAction]

	target := strings.Split(appState.SelectedFile, ".")[0]
	if appState.SelectedAction == models.Actions.CollectionDelete {
		target = appState.CurrentSection.Name
	}

	message := fmt.Sprintf("%s %s?", actionMessage, target)

	code, err := commonUI.ShowMessageWithOptions(message, "0",
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

	return shared.ListSelection{ExitCode: code}, nil
}

func renameRomScreen() (shared.ListSelection, error) {
	logger := common.GetLoggerInstance()
	appState := state.GetAppState()

	args := []string{"--initial-value", appState.SelectedFile, "--title", "Rename ROM", "--show-hardware-group"}

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
		return shared.ListSelection{ExitCode: 1}, err
	}

	outValue := stdoutbuf.String()
	_ = stderrbuf.String()

	return shared.ListSelection{Value: strings.TrimSpace(outValue), ExitCode: cmd.ProcessState.ExitCode()}, nil
}

func downloadArtScreen() (shared.ListSelection, error) {
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

	return shared.ListSelection{ExitCode: exitCode}, nil
}
