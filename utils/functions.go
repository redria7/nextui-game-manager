package utils

import (
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"go.uber.org/zap"
	"nextui-game-manager/state"
	"os"
	"path/filepath"
	"strings"
)

func GetFileList(dirPath string) ([]string, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}
	return files, nil
}

func FilterList(itemList []models.Item, keywords ...string) []models.Item {
	var filteredItemList []models.Item

	for _, item := range itemList {
		for _, keyword := range keywords {
			if strings.Contains(strings.ToLower(item.Filename), strings.ToLower(keyword)) {
				filteredItemList = append(filteredItemList, item)
				break
			}
		}
	}

	return filteredItemList
}

func RefreshRomsList() error {
	appState := state.GetAppState()

	romEntries, err := GetFileList(appState.CurrentSection.LocalDirectory)
	if err != nil {
		return err
	}

	var roms models.Items

	for _, entry := range romEntries {
		roms = append(roms, models.Item{
			Filename: entry,
		})
	}

	appState.CurrentItemsList = roms

	state.UpdateAppState(appState)

	return nil
}

func InsertIntoSlice(s []string, index int, values ...string) []string {
	if index < 0 {
		index = 0
	}
	if index > len(s) {
		index = len(s)
	}

	return append(s[:index], append(values, s[index:]...)...)
}

func FindArt() bool {
	logger := common.GetLoggerInstance()
	appState := state.GetAppState()

	tag := common.TagRegex.FindStringSubmatch(appState.CurrentSection.LocalDirectory)

	if tag == nil {
		return false
	}

	client := common.NewThumbnailClient()
	section := client.BuildThumbnailSection(tag[1])

	artList, err := client.ListDirectory(section)

	if err != nil {
		logger.Info("Unable to fetch artlist", zap.Error(err))
		return false
	}

	noExtension := strings.TrimSuffix(appState.SelectedFile, filepath.Ext(appState.SelectedFile))

	var matched models.Item

	// naive search first
	for _, art := range artList {
		if strings.Contains(strings.ToLower(art.Filename), strings.ToLower(noExtension)) {
			matched = art
			break
		}
	}

	if matched.Filename == "" {
		// TODO Levenshtein Distance support at some point
	}

	if matched.Filename != "" {
		lastSavedArtPath, err := client.DownloadFileRename(section.HostSubdirectory,
			filepath.Join(appState.CurrentSection.LocalDirectory, ".media"), matched.Filename, appState.SelectedFile)

		if err != nil {
			return false
		}

		appState.LastSavedArtPath = lastSavedArtPath

		state.UpdateAppState(appState)

		return true
	}

	return false
}

func FindExistingArt() (string, error) {
	logger := common.GetLoggerInstance()
	appState := state.GetAppState()

	artDir := filepath.Join(appState.CurrentSection.LocalDirectory, ".media")
	artList, err := GetFileList(artDir)
	if err != nil {
		logger.Fatal("failed to list arts", zap.Error(err))
		return "", err
	}

	artFilename := ""

	for _, art := range artList {
		if strings.Contains(art, appState.SelectedFile) {
			artFilename = art
			break
		}
	}

	return artFilename, err
}

func RenameRom(filename string) {
	logger := common.GetLoggerInstance()
	appState := state.GetAppState()

	selectedFile := appState.CurrentItemListWithExtensionMap[appState.SelectedFile]

	oldPath := filepath.Join(appState.CurrentSection.LocalDirectory, selectedFile)
	oldExt := filepath.Ext(selectedFile)
	newPath := filepath.Join(appState.CurrentSection.LocalDirectory, filename+oldExt)

	logger.Debug("Renaming Rom", zap.String("oldPath", oldPath), zap.String("newPath", newPath))

	err := renameFile(oldPath, newPath)
	if err == nil {
		existingArtFilename, err := FindExistingArt()
		if err != nil {
			logger.Error("failed to find existing art", zap.Error(err))
		} else {
			oldArtPath := filepath.Join(appState.CurrentSection.LocalDirectory, ".media", existingArtFilename)
			oldArtExt := filepath.Ext(existingArtFilename)
			newArtPath := filepath.Join(appState.CurrentSection.LocalDirectory, ".media", filename+oldArtExt)

			err := renameFile(oldArtPath, newArtPath)
			if err != nil {
				logger.Error("failed to rename existing art", zap.Error(err))
			}
		}

		appState.SelectedFile = filename
		state.UpdateAppState(appState)

		err = RefreshRomsList()
		if err != nil {
			logger.Error("failed to refresh roms", zap.Error(err))
		}
	}

}

func DeleteRom() {

}

func Nuke() {

}

func ClearGameTracker() {

}

func renameFile(oldPath, newPath string) error {
	logger := common.GetLoggerInstance()

	err := os.Rename(oldPath, newPath)
	if err != nil {
		logger.Error("Failed to rename file", zap.Error(err))
		return err
	}

	return nil
}
