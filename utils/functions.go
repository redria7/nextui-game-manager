package utils

import (
	"bufio"
	"database/sql"
	"fmt"
	gaba "github.com/UncleJunVIP/gabagool/pkg/gabagool"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/filebrowser"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"github.com/disintegration/imaging"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"nextui-game-manager/models"
	"os"
	"path/filepath"
	"qlova.tech/sum"
	"maps"
	"slices"
	"strings"
	"strconv"
	"sync"
	"time"
)

const (
	gameTrackerDBPath = "/mnt/SDCARD/.userdata/shared/game_logs.sqlite"
	saveFileDirectory = "/mnt/SDCARD/Saves/"
	defaultDirPerm    = 0755
	defaultFilePerm   = 0644
)

func IsDev() bool {
	return os.Getenv("ENVIRONMENT") == "DEV"
}

func GetRomDirectory() string {
	if IsDev() {
		return os.Getenv("ROM_DIRECTORY")
	}
	return common.RomDirectory
}

func GetArchiveRoot(archiveName string) string {
	if !strings.HasPrefix(archiveName, ".") {
		archiveName = "." + archiveName
	}

	return filepath.Join(GetRomDirectory(), archiveName)
}

func GetCollectionDirectory() string {
	dir := common.CollectionDirectory
	if IsDev() {
		dir = os.Getenv("COLLECTION_DIRECTORY")
	}

	_ = EnsureDirectoryExists(dir)
	return dir
}

func GetSaveFileDirectory() string {
	if IsDev() {
		return os.Getenv("SAVE_FILE_DIRECTORY")
	}
	return saveFileDirectory
}

func GetGameTrackerDBPath() string {
	if IsDev() {
		return os.Getenv("GAME_TRACKER_DB_PATH")
	}
	return gameTrackerDBPath
}

func GetFileList(dirPath string) ([]os.DirEntry, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", dirPath, err)
	}
	return entries, nil
}

func GetArchiveFileListBasic() ([]string, error) {
	entries, err := GetFileList(GetRomDirectory())
	if err != nil {
		return nil, err
	}

	var archiveFolders []string
	for _, folder := range entries {
		folderName := folder.Name()
		if directoryExists(filepath.Join(GetRomDirectory(), folderName)) {
			if strings.HasPrefix(folderName, ".") {
				if folderName != ".media" {
					archiveFolders = append(archiveFolders, folderName)
				}
			}
		}
	}

	return archiveFolders, nil
}

func GetArchiveFileList() ([]string, error) {
	archiveFolders, err := GetArchiveFileListBasic()

	if err != nil {
		return nil, err
	}

	if archiveFolders == nil {
		archiveFolders = append(archiveFolders, ".Archive")
	}
	return archiveFolders, nil
}

func EnsureDirectoryExists(dirPath string) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return os.MkdirAll(dirPath, defaultDirPerm)
	}
	return nil
}

func CreateRomDirectoryFromItem(item shared.Item) shared.RomDirectory {
	return shared.RomDirectory{
		DisplayName: item.DisplayName,
		Tag:         item.Tag,
		Path:        item.Path,
	}
}

func MoveFile(sourcePath, destinationPath string) error {
	logger := common.GetLoggerInstance()

	if err := EnsureDirectoryExists(filepath.Dir(destinationPath)); err != nil {
		logger.Error("Failed to create destination directory", zap.Error(err))
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	if err := os.Rename(sourcePath, destinationPath); err != nil {
		logger.Error("Failed to move file", zap.String("from", sourcePath), zap.String("to", destinationPath), zap.Error(err))
		return fmt.Errorf("failed to move file from %s to %s: %w", sourcePath, destinationPath, err)
	}

	return nil
}

func FilterList(itemList []shared.Item, keywords ...string) []shared.Item {
	if len(keywords) == 0 {
		return itemList
	}

	var filteredItems []shared.Item
	for _, item := range itemList {
		if matchesAnyKeyword(item.Filename, keywords) {
			filteredItems = append(filteredItems, item)
		}
	}
	return filteredItems
}

func FilterPlayList(itemList []models.PlayTrackingAggregate, keywords ...string) []models.PlayTrackingAggregate {
	if len(keywords) == 0 {
		return itemList
	}

	var filteredItems []models.PlayTrackingAggregate
	for _, item := range itemList {
		if matchesAnyKeyword(item.Name, keywords) {
			filteredItems = append(filteredItems, item)
		}
	}
	return filteredItems
}

func matchesAnyKeyword(filename string, keywords []string) bool {
	lowerFilename := strings.ToLower(filename)
	for _, keyword := range keywords {
		if strings.Contains(lowerFilename, strings.ToLower(keyword)) {
			return true
		}
	}
	return false
}

func InsertIntoSlice(slice []string, index int, values ...string) []string {
	if index < 0 {
		index = 0
	}
	if index > len(slice) {
		index = len(slice)
	}

	return append(slice[:index], append(values, slice[index:]...)...)
}

func FindExistingArt(selectedFile string, romDirectory shared.RomDirectory) (string, error) {
	logger := common.GetLoggerInstance()

	mediaDir := filepath.Join(romDirectory.Path, ".media")
	if !directoryExists(mediaDir) {
		logger.Info("No media directory found", zap.String("directory", romDirectory.Path))
		return "", nil
	}

	artList, err := GetFileList(mediaDir)
	if err != nil {
		return "", fmt.Errorf("failed to list art files: %w", err)
	}

	targetName := removeFileExtension(selectedFile)
	for _, art := range artList {
		if removeFileExtension(art.Name()) == targetName {
			return filepath.Join(mediaDir, art.Name()), nil
		}
	}

	return "", nil
}

func FindAllArt(romDirectory shared.RomDirectory, games shared.Items, downloadType sum.Int[shared.ArtDownloadType]) map[shared.Item]string {
	logger := common.GetLoggerInstance()

	artMap := make(map[shared.Item]string)

	client := common.NewThumbnailClient(downloadType)
	section := client.BuildThumbnailSection(cleanTag(romDirectory.Tag))

	artList, err := client.ListDirectory(section.HostSubdirectory)
	if err != nil {
		logger.Info("Unable to fetch art list", zap.Error(err))
		return artMap
	}

	for _, game := range games {
		matchedArt := findMatchingArt(artList, game.Filename)
		if matchedArt.Filename != "" {
			artMap[game] = matchedArt.Filename
		}
	}

	downloadedArtMap := downloadArtConcurrently(artMap, client, section, logger)

	return downloadedArtMap
}

func downloadArtConcurrently(artMap map[shared.Item]string, client *common.ThumbnailClient, section shared.Section, logger *zap.Logger) map[shared.Item]string {
	downloadedArtMap := make(map[shared.Item]string)
	var mu sync.Mutex
	var wg sync.WaitGroup

	maxConcurrency := 10
	semaphore := make(chan struct{}, maxConcurrency)

	for game, art := range artMap {
		wg.Add(1)
		go func(game shared.Item, art string) {
			defer wg.Done()

			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			path, err := client.DownloadArt(section.HostSubdirectory, buildArtDirectory(game), art, game.Filename)
			if err != nil {
				logger.Error("Failed to download art", zap.Error(err))
				return
			}

			mu.Lock()
			downloadedArtMap[game] = path
			mu.Unlock()
		}(game, art)
	}

	wg.Wait()
	return downloadedArtMap
}

func FindArt(romDirectory shared.RomDirectory, game shared.Item, downloadType sum.Int[shared.ArtDownloadType]) string {
	logger := common.GetLoggerInstance()

	artDirectory := buildArtDirectory(game)
	client := common.NewThumbnailClient(downloadType)
	section := client.BuildThumbnailSection(cleanTag(romDirectory.Tag))

	artList, err := client.ListDirectory(section.HostSubdirectory)
	if err != nil {
		logger.Info("Unable to fetch art list", zap.Error(err))
		return ""
	}

	matchedArt := findMatchingArt(artList, game.Filename)
	if matchedArt.Filename == "" {
		return ""
	}

	lastSavedArtPath, err := client.DownloadArt(section.HostSubdirectory, artDirectory, matchedArt.Filename, game.Filename)
	if err != nil {
		return ""
	}

	src, err := imaging.Open(lastSavedArtPath)
	if err != nil {
		logger.Error("Unable to open last saved art", zap.Error(err))
		return ""
	}

	dst := imaging.Resize(src, 500, 0, imaging.Lanczos)

	err = imaging.Save(dst, lastSavedArtPath)
	if err != nil {
		logger.Error("Unable to save resized last saved art", zap.Error(err))
		return ""
	}

	return lastSavedArtPath
}

func FindRomsWithoutArt() (map[shared.RomDirectory][]shared.Item, error) {
	logger := common.GetLoggerInstance()
	romDirectories := make(map[shared.RomDirectory][]shared.Item)

	fb := filebrowser.NewFileBrowser(logger)

	err := fb.CWD(GetRomDirectory(), false)
	if err != nil {
		logger.Error("Failed to get rom directories", zap.Error(err))
		return nil, fmt.Errorf("failed to get rom directories: %w", err)
	}

	for _, dir := range fb.Items {
		romDir := CreateRomDirectoryFromItem(dir)

		if romDir.Tag == "(PORTS)" {
			continue
		}

		romsWithoutArt, err := findRomsWithoutArtInDirectory(romDir)
		if err != nil {
			logger.Error("Failed to process rom directory", zap.String("directory", romDir.Path), zap.Error(err))
			continue
		}

		if len(romsWithoutArt) > 0 {
			romDirectories[romDir] = romsWithoutArt
		}
	}

	return romDirectories, nil
}

func findRomsWithoutArtInDirectory(romDir shared.RomDirectory) ([]shared.Item, error) {
	romFiles, err := getRomFilesRecursive(romDir.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to get ROM files: %w", err)
	}

	var romsWithoutArt []shared.Item
	for _, romFile := range romFiles {
		romNameWithoutExt := removeFileExtension(romFile.Filename)

		artFilename := filepath.Join(filepath.Dir(romFile.Path), ".media", romNameWithoutExt+".png")

		if !fileExists(artFilename) {
			romsWithoutArt = append(romsWithoutArt, romFile)
		}
	}

	return romsWithoutArt, nil
}

func getRomFilesRecursive(dirPath string) ([]shared.Item, error) {
	var romFiles []shared.Item

	fb := filebrowser.NewFileBrowser(common.GetLoggerInstance())
	err := fb.CWDDepth(dirPath, true, -1)

	if err != nil {
		return nil, fmt.Errorf("failed to get rom files: %w", err)
	}

	var selfContainedPaths []string

	for _, item := range fb.Items {
		if strings.Contains(item.Path, ".media") {
			continue
		}

		parent := filepath.Dir(item.Path)
		if slices.Contains(selfContainedPaths, parent) {
			continue
		}

		if item.IsSelfContainedDirectory {
			selfContainedPaths = append(selfContainedPaths, item.Path)
			romFiles = append(romFiles, item)
		}

		if !item.IsDirectory {
			romFiles = append(romFiles, item)
		}
	}

	return romFiles, err
}

func buildArtDirectory(game shared.Item) string {
	romDirectoryPath := filepath.Dir(game.Path)

	if IsDev() {
		adjustedPath := strings.ReplaceAll(romDirectoryPath, common.RomDirectory, GetRomDirectory())
		return filepath.Join(adjustedPath, ".media")
	}
	return filepath.Join(romDirectoryPath, ".media")
}

func cleanTag(tag string) string {
	cleaned := strings.ReplaceAll(tag, "(", "")
	return strings.ReplaceAll(cleaned, ")", "")
}

func findMatchingArt(artList []shared.Item, filename string) shared.Item {
	targetName := strings.ToLower(removeFileExtension(filename))

	// toastd's trick for Libretro Thumbnail Naming
	cleanedName := strings.ReplaceAll(targetName, "&", "_")

	for _, art := range artList {
		if strings.Contains(strings.ToLower(art.Filename), cleanedName) {
			return art
		}
	}
	return shared.Item{}
}

func DeleteArt(filename string, romDirectory shared.RomDirectory) {
	logger := common.GetLoggerInstance()

	artPath, err := FindExistingArt(filename, romDirectory)
	if err != nil {
		logger.Error("Failed to find existing art", zap.Error(err))
		return
	}

	if artPath == "" {
		logger.Info("No art found to delete")
		return
	}

	common.DeleteFile(artPath)
}

func RenameRom(game shared.Item, newFilename string, romDirectory shared.RomDirectory) (string, error) {
	logger := common.GetLoggerInstance()

	oldPath := filepath.Join(romDirectory.Path, game.Filename)
	newPath := buildNewRomPath(romDirectory.Path, newFilename, game.Filename)

	logger.Debug("Renaming ROM", zap.String("from", oldPath), zap.String("to", newPath))

	if err := MoveFile(oldPath, newPath); err != nil {
		return "", fmt.Errorf("failed to rename ROM file: %w", err)
	}

	renameAssociatedFile(game.Filename, newFilename, newPath, ".cue")
	renameAssociatedFile(game.Filename, newFilename, newPath, ".m3u")

	updateGameTrackerForRename(game.Filename, newFilename, romDirectory, logger)
	renameSaveFile(game.Filename, newFilename, romDirectory)
	// renameCollectionEntries(game, game.Filename, romDirectory) TODO need to finish this functionality
	renameArtFile(game.Filename, newFilename, romDirectory, logger)

	return filepath.Base(newPath), nil
}

func buildNewRomPath(romDirectoryPath, newFilename, oldFilename string) string {
	ext := filepath.Ext(oldFilename)
	return filepath.Join(romDirectoryPath, newFilename+ext)
}

func renameAssociatedFile(oldFilename string, newFilename string, newPath string, extension string) {
	logger := common.GetLoggerInstance()

	oldAssociatedFilename := removeFileExtension(oldFilename) + extension
	oldAssociatedPath := filepath.Join(newPath, oldAssociatedFilename)

	if !fileExists(oldAssociatedPath) {
		return
	}

	newAssociatedFilename := newFilename + extension
	newAssociatedPath := filepath.Join(newPath, newAssociatedFilename)

	if err := MoveFile(oldAssociatedPath, newAssociatedPath); err != nil {
		logger.Error("Failed to rename associated file",
			zap.String("from", oldAssociatedPath),
			zap.String("to", newAssociatedPath),
			zap.String("extension", extension),
			zap.Error(err))
	}
}

func updateGameTrackerForRename(oldFilename, newFilename string, romDirectory shared.RomDirectory, logger *zap.Logger) {
	oldPath := buildGameTrackerPath(romDirectory.Path, oldFilename)
	newPath := buildGameTrackerPath(romDirectory.Path, newFilename+filepath.Ext(oldFilename))

	logger.Debug("Updating game tracker for rename", zap.String("old", oldPath), zap.String("new", newPath))
	MigrateGameTrackerData(newFilename, oldPath, newPath)
}

func buildGameTrackerPath(romDirectoryPath, filename string) string {
	fullPath := filepath.Join(romDirectoryPath, filename)
	return strings.ReplaceAll(fullPath, common.RomDirectory+"/", "")
}

func renameArtFile(oldFilename, newFilename string, romDirectory shared.RomDirectory, logger *zap.Logger) {
	existingArtPath, err := FindExistingArt(oldFilename, romDirectory)
	if err != nil {
		logger.Error("Failed to find existing art", zap.Error(err))
		return
	}

	if existingArtPath == "" {
		return
	}

	if !fileExists(existingArtPath) {
		logger.Info("Art file does not exist, skipping rename")
		return
	}

	newArtPath := buildNewArtPath(existingArtPath, newFilename)
	if err := MoveFile(existingArtPath, newArtPath); err != nil {
		logger.Error("Failed to rename art file", zap.Error(err))
	}
}

func buildNewArtPath(oldArtPath, newFilename string) string {
	dir := filepath.Dir(oldArtPath)
	ext := filepath.Ext(oldArtPath)
	return filepath.Join(dir, newFilename+ext)
}

func ArchiveRom(selectedGame shared.Item, romDirectory shared.RomDirectory, archiveName string) error {
	logger := common.GetLoggerInstance()

	sourcePath := filepath.Join(romDirectory.Path, selectedGame.Filename)
	destinationPath := buildArchivePath(selectedGame.Filename, romDirectory, archiveName)

	logger.Debug("Archiving ROM", zap.String("from", sourcePath), zap.String("to", destinationPath))

	if err := MoveFile(sourcePath, destinationPath); err != nil {
		return fmt.Errorf("failed to archive ROM: %w", err)
	}

	archiveArtFile(selectedGame.Filename, romDirectory, archiveName, logger)
	return nil
}

func RestoreRom(selectedGame shared.Item, romDirectory shared.RomDirectory, archive shared.RomDirectory) error {
	logger := common.GetLoggerInstance()

	sourcePath := filepath.Join(romDirectory.Path, selectedGame.Filename)
	destinationPath := buildRestorePath(selectedGame.Filename, romDirectory, archive)

	logger.Debug("Restoring ROM", zap.String("from", sourcePath), zap.String("to", destinationPath))

	if err := MoveFile(sourcePath, destinationPath); err != nil {
		return fmt.Errorf("failed to restore ROM: %w", err)
	}

	restoreArtFile(selectedGame.Filename, romDirectory, archive, logger)
	return nil
}

func buildArchivePath(filename string, romDirectory shared.RomDirectory, archiveName string) string {
	archiveRoot := GetArchiveRoot(archiveName)
	subdirectory := strings.ReplaceAll(romDirectory.Path, GetRomDirectory(), "")
	return filepath.Join(archiveRoot, subdirectory, filename)
}

func buildRestorePath(filename string, romDirectory shared.RomDirectory, archive shared.RomDirectory) string {
	subdirectory := strings.ReplaceAll(romDirectory.Path, archive.Path, "")
	return filepath.Join(GetRomDirectory(), subdirectory, filename)
}

func archiveArtFile(filename string, romDirectory shared.RomDirectory, archiveName string, logger *zap.Logger) {
	artPath, err := FindExistingArt(filename, romDirectory)
	if err != nil || artPath == "" {
		return
	}

	archiveRoot := GetArchiveRoot(archiveName)
	subdirectory := strings.ReplaceAll(romDirectory.Path, GetRomDirectory(), "")
	destinationPath := filepath.Join(archiveRoot, subdirectory, ".media", filepath.Base(artPath))

	if err := MoveFile(artPath, destinationPath); err != nil {
		logger.Error("Failed to archive art file", zap.Error(err))
	}
}

func restoreArtFile(filename string, romDirectory shared.RomDirectory, archive shared.RomDirectory, logger *zap.Logger) {
	artPath, err := FindExistingArt(filename, romDirectory)
	if err != nil || artPath == "" {
		return
	}

	subdirectory := strings.ReplaceAll(romDirectory.Path, archive.Path, "")
	destinationPath := filepath.Join(GetRomDirectory(), subdirectory, ".media", filepath.Base(artPath))

	if err := MoveFile(artPath, destinationPath); err != nil {
		logger.Error("Failed to restore art file", zap.Error(err))
	}
}

func DeleteRom(game shared.Item, romDirectory shared.RomDirectory) {
	romPath := filepath.Join(romDirectory.Path, game.Filename)
	if common.DeleteFile(romPath) {
		DeleteArt(game.Filename, romDirectory)
	}
}

func Nuke(game shared.Item, romDirectory shared.RomDirectory) {
	ClearGameTracker(game.Filename, romDirectory)
	DeleteRom(game, romDirectory)
}

func CollectGameAggregateFromGame(gameItem shared.Item, gamePlayMap map[string][]models.PlayTrackingAggregate) (models.PlayTrackingAggregate, string) {
	console := extractItemConsoleName(gameItem)
	playTrackingList := gamePlayMap[console]

	for _, gameAggregate := range playTrackingList {
		if gameAggregate.Name == gameItem.DisplayName {
			return gameAggregate, console
		}
	}

	return models.PlayTrackingAggregate{}, console
}

func convertIntListToStringList(intList []int) []string {
	var stringList []string
	for _, i := range intList {
		stringList = append(stringList, strconv.Itoa(i))
	}
	return stringList
}

func GenerateSingleGameGranularRecords(romIds []int) []models.PlayTrackingGranular {
	if len(romIds) == 0 {
		return nil
	}
	
	logger := common.GetLoggerInstance()
	db, err := openGameTrackerDB()
	if err != nil {
		return nil
	}
	defer closeDB(db)

	romIdString := strings.Join(convertIntListToStringList(romIds), "','")

	rows, err := db.Query("SELECT play_time, created_at, updated_at " +
        				  "FROM play_activity " +
						  "WHERE rom_id in ('"+romIdString+"') " +
        				  "ORDER BY created_at")
	defer rows.Close()

	var granularList []models.PlayTrackingGranular
	for rows.Next() {
		var playTime 	int
		var createTime 	int
		var updateTime 	int
		if err := rows.Scan(&playTime, &createTime, &updateTime); err != nil {
			logger.Error("Failed to load game tracker data", zap.Error(err))
		}

		playTrack := models.PlayTrackingGranular{
			PlayTime:	playTime,
			StartTime: 	createTime,
			UpdateTime:	updateTime,
		}
		granularList = append(granularList, playTrack)
	}

	return granularList
}

func GenerateCurrentGameStats() (map[string][]models.PlayTrackingAggregate, map[string]int, int) {
	logger := common.GetLoggerInstance()
	db, err := openGameTrackerDB()
	if err != nil {
		return nil, nil, 0
	}
	defer closeDB(db)

	rows, err := db.Query("SELECT rom.id, rom.name, rom.file_path, " +
						  "SUM(play_activity.play_time) AS play_time_total, " +
						  "COUNT(play_activity.ROWID) AS play_count_total, " +
						  "MIN(play_activity.created_at) AS first_played_at, " +
						  "MAX(play_activity.created_at) AS last_played_at " +
        				  "FROM rom " +
						  "LEFT JOIN play_activity " +
						  "ON rom.id = play_activity.rom_id " +
        				  "GROUP BY rom.id " +
        				  "HAVING play_time_total > 0 " +
        				  "ORDER BY play_time_total DESC")
	defer rows.Close()

	var gamePlayMap map[string][]models.PlayTrackingAggregate
	var consolePlayMap map[string]int
	totalPlay := 0
	var multiMap map[string]bool
	for rows.Next() {
		var id 				int
		var name 			string
		var filePath 		string
		var playTimeTotal 	int
		var playCountTotal 	int
		var firstPlayedTime int
		var lastPlayedTime 	int
		if err := rows.Scan(&id, &name, &filePath, &playTimeTotal, &playCountTotal, &firstPlayedTime, &lastPlayedTime); err != nil {
			logger.Error("Failed to load game tracker data", zap.Error(err))
		}

		logger.Info("loading", zap.String("extracting name", name))
		romName, multi := extractMultiDiscName(name, filePath)
		logger.Info("loading", zap.String("extracted name", romName))
		playTrack := models.PlayTrackingAggregate{
			Id:					[]int{id},
			Name: 				romName,
			PlayTimeTotal:    	playTimeTotal,
			PlayCountTotal:    	playCountTotal,
			FirstPlayedTime: 	time.Unix(int64(firstPlayedTime), 0),
			LastPlayedTime:    	time.Unix(int64(lastPlayedTime), 0),
		}
		console := extractPlayConsoleName(filePath)
		logger.Info("loading", zap.String("extracted console", console))

		if multi {
			multiMap[console] = true
			gamePlayMap[console] = appendMultiDiscAggregate(gamePlayMap[console], playTrack)
		} else {
			gamePlayMap[console] = append(gamePlayMap[console], playTrack)
		}
		logger.Info("loading", zap.String("mapped game play", strconv.Itoa(gamePlayMap[console][0].Id[0])))

		consolePlayMap[console] = consolePlayMap[console] + playTrack.PlayTimeTotal
		logger.Info("loading", zap.String("mapped console play", strconv.Itoa(consolePlayMap[console])))

		totalPlay = totalPlay + playTrack.PlayTimeTotal
		logger.Info("loading", zap.String("mapped total play", strconv.Itoa(totalPlay)))
	}

	gamePlayMap = sortPlayMap(gamePlayMap, multiMap)

	return gamePlayMap, consolePlayMap, totalPlay
}

func sortPlayMap(playMap map[string][]models.PlayTrackingAggregate, multiMap map[string]bool) map[string][]models.PlayTrackingAggregate {
	keys := slices.Sorted(maps.Keys(multiMap))
	for _, key := range keys {
		aggregateList := playMap[key]
		slices.SortFunc(aggregateList, func(a, b models.PlayTrackingAggregate) int {
			return a.PlayTimeTotal - b.PlayTimeTotal
		})
		playMap[key] = aggregateList
	}
	return playMap
}

func appendMultiDiscAggregate(existingList []models.PlayTrackingAggregate, newAggregate models.PlayTrackingAggregate) []models.PlayTrackingAggregate {
	for index, existingAggregate := range existingList {
		if existingAggregate.Name == newAggregate.Name {
			existingList[index] = models.PlayTrackingAggregate{
				Id:					appendUniqueAggregateId(existingAggregate.Id, newAggregate.Id[0]),
				Name: 				existingAggregate.Name,
				PlayTimeTotal:    	existingAggregate.PlayTimeTotal+newAggregate.PlayTimeTotal,
				PlayCountTotal:    	existingAggregate.PlayCountTotal+newAggregate.PlayCountTotal,
				FirstPlayedTime: 	minTime(existingAggregate.FirstPlayedTime, newAggregate.FirstPlayedTime),
				LastPlayedTime:    	maxTime(existingAggregate.FirstPlayedTime, newAggregate.FirstPlayedTime),
			}
			return existingList
		}
	}
	return append(existingList, newAggregate)
}

func appendUniqueAggregateId(existingIds []int, newId int) []int {
	for _, id := range existingIds {
		if id == newId {
			return existingIds
		}
	}
	return append(existingIds, newId)
}

func minTime(a time.Time, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}

func maxTime(a time.Time, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

func extractMultiDiscName(romName string, filePath string) (string, bool) {
	if strings.Contains(romName, "(Disc") || strings.Contains(romName, "(Disk") {
		pathList := strings.Split(filePath, "/")
		return pathList[len(pathList)-2], true
	}
	return romName, false
}

func extractPlayConsoleName(romFilePath string) string {
	return strings.Split(romFilePath, "/")[0]
}

func extractItemConsoleName(gameItem shared.Item) string {
	pathSplit := strings.Split(gameItem.Path, "/")
	if len(pathSplit) < 5 {
		return ""
	}
	return pathSplit[4]
}

func ConvertSecondsToHumanReadable(gameTimeSeconds int) string {
	hours := gameTimeSeconds/3600
	minutes := (gameTimeSeconds/60)%60
	seconds := gameTimeSeconds%60
	return fmt.Sprintf("%d Hours, %d Minutes, %d Seconds", hours, minutes, seconds)
}

func ConvertSecondsToHumanReadableAbbreviated(gameTimeSeconds int) string {
	hours := gameTimeSeconds/3600
	minutes := (gameTimeSeconds/60)%60
	seconds := gameTimeSeconds%60
	return fmt.Sprintf("%dH %dM %dS", hours, minutes, seconds)
}

func HasGameTrackerData(romFilename string, romDirectory shared.RomDirectory) bool {
	db, err := openGameTrackerDB()
	if err != nil {
		return false
	}
	defer closeDB(db)

	romPath := buildGameTrackerPath(romDirectory.Path, romFilename)

	var romID string
	err = db.QueryRow("SELECT id FROM rom WHERE file_path = ?", romPath).Scan(&romID)
	return err == nil && romID != ""
}

func MigrateGameTrackerData(filename, oldPath, newPath string) bool {
	logger := common.GetLoggerInstance()

	db, err := openGameTrackerDB()
	if err != nil {
		return false
	}
	defer closeDB(db)

	logger.Debug("Migrating game tracker data",
		zap.String("filename", filename),
		zap.String("oldPath", oldPath),
		zap.String("newPath", newPath))

	return executeGameTrackerMigration(db, filename, oldPath, newPath, logger)
}

func executeGameTrackerMigration(db *sql.DB, filename, oldPath, newPath string, logger *zap.Logger) bool {
	tx, err := db.Begin()
	if err != nil {
		logger.Error("Failed to begin transaction", zap.Error(err))
		return false
	}
	defer tx.Rollback()

	romID, err := findRomID(tx, oldPath)
	if err != nil || romID == "" {
		logger.Error("Failed to find ROM ID", zap.String("path", oldPath), zap.Error(err))
		return false
	}

	if err := updateRomData(tx, filename, newPath, romID); err != nil {
		logger.Error("Failed to update ROM data", zap.Error(err))
		return false
	}

	if err := tx.Commit(); err != nil {
		logger.Error("Failed to commit transaction", zap.Error(err))
		return false
	}

	logger.Info("Game tracker ROM data updated successfully")
	return true
}

func ClearGameTracker(romName string, romDirectory shared.RomDirectory) bool {
	logger := common.GetLoggerInstance()

	db, err := openGameTrackerDB()
	if err != nil {
		return false
	}
	defer closeDB(db)

	romPath := buildGameTrackerPath(romDirectory.Path, romName)
	return executeGameTrackerClear(db, romPath, logger)
}

func executeGameTrackerClear(db *sql.DB, romPath string, logger *zap.Logger) bool {
	tx, err := db.Begin()
	if err != nil {
		logger.Error("Failed to begin transaction", zap.Error(err))
		return false
	}
	defer tx.Rollback()

	romID, err := findRomID(tx, romPath)
	if err != nil || romID == "" {
		logger.Warn("No ROM found to clear", zap.String("path", romPath))
		return false
	}

	if err := deleteGameTrackerData(tx, romID); err != nil {
		logger.Error("Failed to delete game tracker data", zap.Error(err))
		return false
	}

	if err := tx.Commit(); err != nil {
		logger.Error("Failed to commit transaction", zap.Error(err))
		return false
	}

	logger.Info("Game tracker data cleared successfully")
	return true
}

func openGameTrackerDB() (*sql.DB, error) {
	logger := common.GetLoggerInstance()

	db, err := sql.Open("sqlite3", GetGameTrackerDBPath())
	if err != nil {
		logger.Error("Failed to open game tracker database", zap.Error(err))
		return nil, err
	}
	return db, nil
}

func closeDB(db *sql.DB) {
	if err := db.Close(); err != nil {
		logger := common.GetLoggerInstance()
		logger.Error("Failed to close database", zap.Error(err))
	}
}

func findRomID(tx *sql.Tx, romPath string) (string, error) {
	var romID string
	err := tx.QueryRow("SELECT id FROM rom WHERE file_path = ?", romPath).Scan(&romID)
	return romID, err
}

func updateRomData(tx *sql.Tx, filename, newPath, romID string) error {
	_, err := tx.Exec("UPDATE rom SET name = ?, file_path = ? WHERE id = ?", filename, newPath, romID)
	return err
}

func deleteGameTrackerData(tx *sql.Tx, romID string) error {
	if _, err := tx.Exec("DELETE FROM play_activity WHERE rom_id = ?", romID); err != nil {
		return err
	}
	_, err := tx.Exec("DELETE FROM rom WHERE id = ?", romID)
	return err
}

func renameSaveFile(oldFilename, newFilename string, romDirectory shared.RomDirectory) {
	logger := common.GetLoggerInstance()

	saveDir := buildSaveDirectory(romDirectory)
	fb := filebrowser.NewFileBrowser(logger)

	if err := fb.CWD(saveDir, true); err != nil {
		logger.Error("Failed to access save directory", zap.String("dir", saveDir), zap.Error(err))
		return
	}

	saveFile := findSaveFile(fb.Items, oldFilename)
	if saveFile.Filename == "" {
		logger.Info("No save file found to rename")
		return
	}

	newSavePath := buildNewSavePath(saveDir, oldFilename, newFilename, saveFile.Filename)
	if err := MoveFile(saveFile.Path, newSavePath); err != nil {
		logger.Error("Failed to rename save file", zap.Error(err))
	}
}

func buildSaveDirectory(romDirectory shared.RomDirectory) string {
	tag := cleanTag(romDirectory.Tag)
	return filepath.Join(GetSaveFileDirectory(), tag)
}

func findSaveFile(items []shared.Item, targetFilename string) shared.Item {
	lowerTarget := strings.ToLower(targetFilename)
	for _, item := range items {
		if strings.Contains(strings.ToLower(item.Filename), lowerTarget) {
			return item
		}
	}
	return shared.Item{}
}

func buildNewSavePath(saveDir, oldFilename, newFilename, saveFilename string) string {
	ext := strings.ReplaceAll(saveFilename, removeFileExtension(oldFilename), "")
	return filepath.Join(saveDir, newFilename+ext)
}

func renameCollectionEntries(game shared.Item, oldDisplayName string, romDirectory shared.RomDirectory) {
	logger := common.GetLoggerInstance()

	collections := findCollectionsContainingGame(game, logger)
	for _, collection := range collections {
		updateCollectionGamePath(collection, oldDisplayName, game, romDirectory)
		SaveCollection(collection)
	}
}

func findCollectionsContainingGame(game shared.Item, logger *zap.Logger) []models.Collection {
	fb := filebrowser.NewFileBrowser(logger)
	if err := fb.CWD(GetCollectionDirectory(), false); err != nil {
		return nil
	}

	var collections []models.Collection
	for _, item := range fb.Items {
		collection := models.Collection{
			DisplayName:    item.DisplayName,
			CollectionFile: item.Path,
		}

		if loadedCollection, err := ReadCollection(collection); err == nil {
			if containsGame(loadedCollection.Games, game) {
				collections = append(collections, loadedCollection)
			}
		}
	}

	return collections
}

func containsGame(games []shared.Item, targetGame shared.Item) bool {
	return slices.ContainsFunc(games, func(game shared.Item) bool {
		return game.DisplayName == targetGame.DisplayName
	})
}

func updateCollectionGamePath(collection models.Collection, oldDisplayName string, game shared.Item, romDirectory shared.RomDirectory) {
	romDirectoryStub := strings.ReplaceAll(romDirectory.Path, GetRomDirectory(), "")

	for i, item := range collection.Games {
		if item.DisplayName == oldDisplayName {
			collection.Games[i].Path = filepath.Join(romDirectoryStub, game.Filename)
		}
	}
}

func RenameCollection(collection models.Collection, name string) (models.Collection, error) {
	logger := common.GetLoggerInstance()

	newPath := filepath.Join(filepath.Dir(collection.CollectionFile), name+".txt")

	if err := os.Rename(collection.CollectionFile, newPath); err != nil {
		logger.Error("Failed to rename collection file", zap.Error(err))
		return models.Collection{}, fmt.Errorf("failed to rename collection: %w", err)
	}

	collection.DisplayName = name
	collection.CollectionFile = newPath
	return collection, nil
}

func DeleteCollection(collection models.Collection) {
	common.DeleteFile(collection.CollectionFile)
}

func AddCollectionGames(collection models.Collection, games []shared.Item) (models.Collection, error) {
	logger := common.GetLoggerInstance()

	if fileExists(collection.CollectionFile) {
		logger.Debug("Loading existing collection")

		if loadedCollection, err := ReadCollection(collection); err == nil {
			collection = loadedCollection
		} else {
			return collection, fmt.Errorf("failed to load existing collection: %w", err)
		}
	}

	for _, game := range games {
		if GameExistsInCollection(collection.Games, game) {
			logger.Debug("Game already exists in collection", zap.String("path", game.Path))
			continue
		}
		collection.Games = append(collection.Games, game)
	}

	return collection, SaveCollection(collection)
}

func GameExistsInCollection(games []shared.Item, targetGame shared.Item) bool {
	for _, game := range games {
		if strings.Contains(strings.ToLower(game.Path), strings.ToLower(targetGame.DisplayName)) {
			return true
		}
	}
	return false
}

func ReadCollection(collection models.Collection) (models.Collection, error) {
	logger := common.GetLoggerInstance()

	file, err := os.Open(collection.CollectionFile)
	if err != nil {
		logger.Error("Failed to open collection file", zap.String("file", collection.CollectionFile), zap.Error(err))
		return collection, fmt.Errorf("failed to open collection file: %w", err)
	}
	defer file.Close()

	var games []shared.Item
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		displayName := strings.ReplaceAll(filepath.Base(line), filepath.Ext(line), "")

		games = append(games, shared.Item{
			DisplayName: displayName,
			Path:        line,
		})
	}

	if err := scanner.Err(); err != nil {
		logger.Error("Failed to read collection file", zap.Error(err))
		return collection, fmt.Errorf("failed to read collection: %w", err)
	}

	collection.Games = games
	return collection, nil
}

func SaveCollection(collection models.Collection) error {
	if err := EnsureDirectoryExists(filepath.Dir(collection.CollectionFile)); err != nil {
		return fmt.Errorf("failed to create collection directory: %w", err)
	}

	file, err := os.OpenFile(collection.CollectionFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, defaultFilePerm)
	if err != nil {
		return fmt.Errorf("failed to open collection file: %w", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	for _, game := range collection.Games {
		path := normalizeGamePath(game)
		if _, err := writer.WriteString(path + "\n"); err != nil {
			return fmt.Errorf("failed to write collection entry: %w", err)
		}
	}

	return nil
}

func normalizeGamePath(game shared.Item) string {
	path := strings.ReplaceAll(game.Path, GetRomDirectory()+"/", "/Roms/")

	if game.IsMultiDiscDirectory {
		path = filepath.Join(path, game.DisplayName+".m3u")
	}

	return path
}

func SaveConfig(config *models.Config) error {
	configFile := "config.yml"

	if !fileExists(configFile) {
		if err := CreateEmptyConfigFile(configFile); err != nil {
			return fmt.Errorf("failed to create config file: %w", err)
		}
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to setup viper: %w", err)
	}

	viper.Set("art_download_type", config.ArtDownloadType)
	viper.Set("hide_empty", config.HideEmpty)
	viper.Set("show_art", config.ShowArt)
	viper.Set("log_level", config.LogLevel)

	return viper.WriteConfigAs(configFile)
}

func CreateEmptyConfigFile(filename string) error {
	logger := common.GetLoggerInstance()
	logger.Info("Creating new config file", zap.String("file", filename))

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	return file.Close()
}

func removeFileExtension(filename string) string {
	return strings.TrimSuffix(filename, filepath.Ext(filename))
}

func directoryExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func PrepArchiveName(archive string) string {
	if !strings.HasPrefix(archive, ".") {
		return "." + archive
	}
	return archive
}

func CleanArchiveName(archive string) string {
	return strings.TrimPrefix(archive, ".")
}

func DeleteArchive(archive shared.RomDirectory) (string, error) {
	logger := common.GetLoggerInstance()
	res, err := deleteArchiveRecursive(archive.Path, 0)

	if err != nil {
		logger.Error("Failed to traverse archive", zap.Error(err))
		return res, err
	}

	if res != "" {
		return res, nil
	}

	removeErr := os.RemoveAll(archive.Path)

	if removeErr != nil {
		return "", removeErr
	}

	return "", nil
}

func deleteArchiveRecursive(currentDirectory string, currentDepth int) (string, error) {
	logger := common.GetLoggerInstance()
	if currentDepth > 10 {
		return "Max Depth Exceeded", nil
	}

	entries, err := GetFileList(currentDirectory)

	if err != nil {
		logger.Error("Failed to traverse archive", zap.Error(err))
		return "", err
	}

	for _, file := range entries {
		if !file.IsDir() {
			return file.Name(), nil
		}

		res, recurseErr := deleteArchiveRecursive(filepath.Join(currentDirectory, file.Name()), currentDepth+1)

		if recurseErr != nil {
			return "", recurseErr
		}

		if res != "" {
			return res, nil
		}
	}

	return "", nil
}

func ShowTimedMessage(message string, delay time.Duration) {
	gaba.ProcessMessage(message, gaba.ProcessMessageOptions{}, func() (interface{}, error) {
		time.Sleep(delay)
		return nil, nil
	})
}

func ConfirmAction(message string) bool {
	result, err := gaba.ConfirmationMessage(message, []gaba.FooterHelpItem{
		{ButtonName: "B", HelpText: "I Changed My Mind"},
		{ButtonName: "A", HelpText: "Yes"},
	}, gaba.MessageOptions{})

	return err == nil && result.IsSome()
}

func ConfirmBulkAction(message string) bool {
	confirm, _ := gaba.ConfirmationMessage(message, []gaba.FooterHelpItem{
		{ButtonName: "B", HelpText: "Cancel"},
		{ButtonName: "X", HelpText: "Remove"},
	}, gaba.MessageOptions{
		ImagePath:     "",
		ConfirmButton: gaba.ButtonX,
	})

	return confirm.IsSome() && !confirm.Unwrap().Cancelled
}
