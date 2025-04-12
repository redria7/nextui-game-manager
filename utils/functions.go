package utils

import (
	"database/sql"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
	"nextui-game-manager/state"
	"os"
	"path/filepath"
	"strings"
)

const gameTrackerDBPath = "/mnt/SDCARD/.userdata/shared/game_logs.sqlite"

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

	mediaDir := filepath.Join(appState.CurrentSection.LocalDirectory, ".media")

	if _, err := os.Stat(mediaDir); os.IsNotExist(err) {
		logger.Info("No media directory found", zap.String("current_directory", appState.CurrentSection.LocalDirectory))
		return "", nil
	}

	artDir := filepath.Join(appState.CurrentSection.LocalDirectory, ".media")
	artList, err := GetFileList(artDir)
	if err != nil {
		logger.Error("failed to list arts", zap.Error(err))
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

	err := moveFile(oldPath, newPath)
	if err == nil {
		existingArtFilename, err := FindExistingArt()
		if err != nil {
			logger.Error("failed to find existing art", zap.Error(err))
		} else {
			oldArtPath := filepath.Join(appState.CurrentSection.LocalDirectory, ".media", existingArtFilename)
			oldArtExt := filepath.Ext(existingArtFilename)
			newArtPath := filepath.Join(appState.CurrentSection.LocalDirectory, ".media", filename+oldArtExt)

			err := moveFile(oldArtPath, newArtPath)
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

func DeleteArt() {
	logger := common.GetLoggerInstance()
	appState := state.GetAppState()

	art, err := FindExistingArt()
	if err != nil {
		logger.Error("failed to find existing art", zap.Error(err))
		return
	} else if art == "" {
		logger.Info("No art. Skipping delete.")
		return
	}

	artPath := filepath.Join(appState.CurrentSection.LocalDirectory, ".media", art)
	common.DeleteFile(artPath)
}

func HasGameTrackerData() bool {
	logger := common.GetLoggerInstance()
	appState := state.GetAppState()

	db, err := sql.Open("sqlite3", gameTrackerDBPath)
	if err != nil {
		logger.Error("Failed to open game tracker database", zap.Error(err))
		return false
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			logger.Error("Failed to close game tracker database", zap.Error(err))
		}
	}(db)

	tag := common.TagRegex.FindStringSubmatch(appState.CurrentSection.LocalDirectory)
	tagWildCard := "%" + tag[1] + "%"
	romName := appState.SelectedFile // Replace with actual name if needed

	var romID string
	err = db.QueryRow("SELECT id FROM rom WHERE file_path LIKE ? AND name = ?", tagWildCard, romName).Scan(&romID)
	if err != nil {
		logger.Error("Failed to find ROM ID", zap.Error(err))
		return false
	}

	return romID != ""
}

func ClearGameTracker() bool {
	logger := common.GetLoggerInstance()
	appState := state.GetAppState()

	db, err := sql.Open("sqlite3", gameTrackerDBPath)
	if err != nil {
		logger.Error("Failed to open game tracker database", zap.Error(err))
		return false
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			logger.Error("Failed to close game tracker database", zap.Error(err))
		}
	}(db)

	tx, err := db.Begin()
	if err != nil {
		logger.Error("Failed to begin transaction", zap.Error(err))
		return false
	}

	tag := common.TagRegex.FindStringSubmatch(appState.CurrentSection.LocalDirectory)
	tagWildCard := "%" + tag[1] + "%"
	romName := appState.SelectedFile // Replace with actual name if needed

	var romID string
	err = tx.QueryRow("SELECT id FROM rom WHERE file_path LIKE ? AND name = ?", tagWildCard, romName).Scan(&romID)
	if err != nil {
		_ = tx.Rollback()
		logger.Error("Failed to find ROM ID", zap.Error(err))
		return false
	}

	if romID == "" {
		logger.Warn("No ROM ID found", zap.String("tag", tag[1]), zap.String("name", romName))
		return false
	}

	_, err = tx.Exec("DELETE FROM play_activity WHERE rom_id = ?", romID)
	if err != nil {
		_ = tx.Rollback()
		logger.Error("Failed to delete play activity", zap.Error(err))
		return false
	}

	_, err = tx.Exec("DELETE FROM rom WHERE id = ?", romID)
	if err != nil {
		_ = tx.Rollback()
		logger.Error("Failed to delete rom", zap.Error(err))
		return false
	}

	err = tx.Commit()
	if err != nil {
		logger.Error("Failed to commit transaction", zap.Error(err))
		return false
	}

	logger.Info("Game tracker data cleared successfully")
	return true
}

func ClearSaveStates() {

}

func ArchiveRom() {
	const archiveRoot = "/mnt/SDCARD/Roms/.Archive"

	logger := common.GetLoggerInstance()
	appState := state.GetAppState()

	selectedFile := appState.CurrentItemListWithExtensionMap[appState.SelectedFile]

	oldPath := filepath.Join(appState.CurrentSection.LocalDirectory, selectedFile)
	oldPathSubdirectory := strings.ReplaceAll(appState.CurrentSection.LocalDirectory, common.RomDirectory, "")
	newPath := filepath.Join(archiveRoot, oldPathSubdirectory, selectedFile)

	logger.Debug("Archiving Rom", zap.String("oldPath", oldPath), zap.String("newPath", newPath))

	err := moveFile(oldPath, newPath)
	if err == nil {
		existingArtFilename, err := FindExistingArt()
		if err != nil {
			logger.Error("failed to find existing art", zap.Error(err))
		} else {
			oldArtPath := filepath.Join(appState.CurrentSection.LocalDirectory, ".media", existingArtFilename)
			newArtPath := filepath.Join(archiveRoot, oldPathSubdirectory, ".media", existingArtFilename)

			err := moveFile(oldArtPath, newArtPath)
			if err != nil {
				logger.Error("failed to archive existing art", zap.Error(err))
			}
		}

		appState.SelectedFile = ""
		state.UpdateAppState(appState)

		err = RefreshRomsList()
		if err != nil {
			logger.Error("failed to refresh roms", zap.Error(err))
		}
	}
}

func DeleteRom() {
	appState := state.GetAppState()
	filename := appState.CurrentItemListWithExtensionMap[appState.SelectedFile]
	romPath := filepath.Join(appState.CurrentSection.LocalDirectory, filename)
	res := common.DeleteFile(romPath)

	if res {
		appState.SelectedFile = ""
		state.UpdateAppState(appState)
		DeleteArt()
		err := RefreshRomsList()
		if err != nil {
			logger := common.GetLoggerInstance()
			logger.Error("failed to refresh roms", zap.Error(err))
		}
	}
}

func Nuke() {
	ClearGameTracker()
	DeleteArt()
	DeleteRom()
	err := RefreshRomsList()
	if err != nil {
		logger := common.GetLoggerInstance()
		logger.Error("failed to refresh roms", zap.Error(err))
	}
}

func moveFile(oldPath, newPath string) error {
	logger := common.GetLoggerInstance()

	dir := filepath.Dir(newPath)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		logger.Error("Failed to create destination directory", zap.Error(err))
		return err
	}

	err := os.Rename(oldPath, newPath)
	if err != nil {
		logger.Error("Failed to move file", zap.Error(err))
		return err
	}

	return nil
}
