package utils

import (
	"fmt"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/filebrowser"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	_ "github.com/mattn/go-sqlite3"
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
	gameTrackerDBPath  = "/mnt/SDCARD/.userdata/shared/game_logs.sqlite"
	saveFileDirectory  = "/mnt/SDCARD/Saves/"
	RecentlyPlayedFile = "/mnt/SDCARD/.userdata/shared/.minui/recent.txt"
	defaultDirPerm     = 0755
	defaultFilePerm    = 0644
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

func CreateRomDirectoryFromItem(item shared.Item) shared.RomDirectory {
	return shared.RomDirectory{
		DisplayName: item.DisplayName,
		Tag:         item.Tag,
		Path:        item.Path,
	}
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

func GenerateCollectionMap() map[string][]models.Collection {
	collectionMap := make(map[string][]models.Collection)
	collectionList, _, _ := GenerateCollectionList("", false)
	for _, collection := range collectionList {
		for _, game := range collection.Games {
			collectionMap[game.DisplayName] = append(collectionMap[game.DisplayName], collection)
		}
	}
	return collectionMap
}

func GenerateCollectionList(searchFilter string, onScreen bool) (collections []models.Collection, exitCode int, e error) {
	fb := filebrowser.NewFileBrowser(common.GetLoggerInstance())
	err := fb.CWD(GetCollectionDirectory(), false)
	if err != nil {
		if onScreen {
			ShowTimedMessage("Unable to Load Collections!", time.Second*2)
		}
		return nil, 404, nil
	}

	if fb.Items == nil || len(fb.Items) == 0 {
		return nil, 404, nil
	}

	itemList := fb.Items

	if searchFilter != "" {
		itemList = FilterList(itemList, searchFilter)
	}

	slices.SortFunc(itemList, func(a, b shared.Item) int {
		return strings.Compare(a.DisplayName, b.DisplayName)
	})

	var collectionList []models.Collection
	for _, item := range itemList {
		col := models.Collection{DisplayName: item.DisplayName, CollectionFile: item.Path}
		col, err = ReadCollection(col)

		if err != nil {
			if onScreen {
				ShowTimedMessage("Unable to Load Collections!", time.Second*2)
			}
			return nil, -1, err
		}

		collectionList = append(collectionList, col)
	}

	return collectionList, 0, nil
}

func FindRomHomeFromAggregate(gameAggregate models.PlayTrackingAggregate, showArchives bool) string {
	gamePath := gameAggregate.Path
	if directoryExists(gamePath) || fileExists(gamePath) {
		if showArchives {
			return "(+) "
		}
		return ""
	}

	archiveList, err := GetArchiveFileListBasic()
	if err == nil {
		for _, archiveName := range archiveList {
			gameSubPath := strings.ReplaceAll(gamePath, GetRomDirectory(), "")
			archivePath := filepath.Join(GetRomDirectory(), archiveName, gameSubPath)
			if directoryExists(archivePath) || fileExists(archivePath) {
				if showArchives {
					return "(" + string(CleanArchiveName(archiveName)[0]) + ") "
				}
				return ""
			}
		}
	}

	return "(-) "
}

func CollectGameAggregateFromGame(gameItem shared.Item, gamePlayMap map[string][]models.PlayTrackingAggregate) (models.PlayTrackingAggregate, string) {
	console := extractItemConsoleName(gameItem)
	playTrackingList := gamePlayMap[console]

	for _, gameAggregate := range playTrackingList {
		if gameAggregate.Name == gameItem.DisplayName {
			return gameAggregate, console
		}
	}

	return models.PlayTrackingAggregate{
		Name: gameItem.DisplayName,
		PlayTimeTotal: 0,
		PlayCountTotal: 0,
	}, console
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

	gamePlayMap := make(map[string][]models.PlayTrackingAggregate)
	consolePlayMap := make(map[string]int)
	totalPlay := 0
	multiMap := make(map[string]bool)
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

		romName, romPath, multi := extractMultiDiscName(name, filePath)
		playTrack := models.PlayTrackingAggregate{
			Id:					[]int{id},
			Name: 				romName,
			Path:				romPath,
			PlayTimeTotal:    	playTimeTotal,
			PlayCountTotal:    	playCountTotal,
			FirstPlayedTime: 	time.Unix(int64(firstPlayedTime), 0),
			LastPlayedTime:    	time.Unix(int64(lastPlayedTime), 0),
		}
		console := extractPlayConsoleName(filePath)

		if multi {
			multiMap[console] = true
			gamePlayMap[console] = appendMultiDiscAggregate(gamePlayMap[console], playTrack)
		} else {
			gamePlayMap[console] = append(gamePlayMap[console], playTrack)
		}

		consolePlayMap[console] = consolePlayMap[console] + playTrack.PlayTimeTotal

		totalPlay = totalPlay + playTrack.PlayTimeTotal
	}

	gamePlayMap = sortPlayMap(gamePlayMap, multiMap)

	return gamePlayMap, consolePlayMap, totalPlay
}

func sortPlayMap(playMap map[string][]models.PlayTrackingAggregate, multiMap map[string]bool) map[string][]models.PlayTrackingAggregate {
	keys := slices.Sorted(maps.Keys(multiMap))
	for _, key := range keys {
		aggregateList := playMap[key]
		slices.SortFunc(aggregateList, func(a, b models.PlayTrackingAggregate) int {
			return  b.PlayTimeTotal - a.PlayTimeTotal
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

func extractMultiDiscName(romName string, filePath string) (string, string, bool) {
	if (strings.Contains(romName, "(Disc") || strings.Contains(romName, "(Disk")) {
		pathList := strings.Split(filePath, "/")
		if len(pathList) >= 2 {
			return pathList[len(pathList)-2], filepath.Join(GetRomDirectory(), filePath, ".."), true
		}
	}
	return romName, filepath.Join(GetRomDirectory(), filePath), false
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

func findSaveFile(items []shared.Item, targetFilename string) shared.Item {
	lowerTarget := strings.ToLower(targetFilename)
	for _, item := range items {
		if strings.Contains(strings.ToLower(item.Filename), lowerTarget) {
			return item
		}
	}
	return shared.Item{}
}

func cleanTag(tag string) string {
	cleaned := strings.ReplaceAll(tag, "(", "")
	return strings.ReplaceAll(cleaned, ")", "")
}

func removeFileExtension(filename string) string {
	return strings.TrimSuffix(filename, filepath.Ext(filename))
}
