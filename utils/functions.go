package utils

import (
	"context"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	"go.uber.org/zap"
	"mortar/clients"
	"mortar/models"
	"nextui-game-manager/state"
	"os"
	"path/filepath"
	"strings"
)

func DeleteFile(path string) {
	logger := common.GetLoggerInstance()

	err := os.Remove(path)
	if err != nil {
		logger.Error("Issue removing file",
			zap.String("path", path),
			zap.Error(err))
	} else {
		logger.Debug("Removed file", zap.String("path", path))
	}
}

func FindArt() bool {
	logger := common.GetLoggerInstance()
	appState := state.GetAppState()

	tag := common.TagRegex.FindStringSubmatch(appState.CurrentSection.LocalDirectory)

	if tag == nil {
		return false
	}

	client := clients.NewThumbnailClient()
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
		err = client.DownloadFileRename(section.HostSubdirectory,
			filepath.Join(appState.CurrentSection.LocalDirectory, ".media"), matched.Filename, appState.SelectedFile)

		if err != nil {
			return false
		}

		return true
	}

	return false
}

func DownloadFile(cancel context.CancelFunc) error {
	defer cancel()

	logger := common.GetLoggerInstance()
	appState := state.GetAppState()

	client, err := clients.BuildClient(appState.CurrentHost)
	if err != nil {
		return err
	}

	defer func(client models.Client) {
		err := client.Close()
		if err != nil {
			logger.Error("Unable to close client", zap.Error(err))
		}
	}(client)

	var hostSubdirectory string

	return client.DownloadFile(hostSubdirectory,
		appState.CurrentSection.LocalDirectory, appState.SelectedFile)
}
