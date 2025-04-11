package utils

import (
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"go.uber.org/zap"
	"nextui-game-manager/state"
	"path/filepath"
	"strings"
)

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
