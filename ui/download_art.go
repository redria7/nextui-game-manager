package ui

import (
	"context"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	cui "github.com/UncleJunVIP/nextui-pak-shared-functions/ui"
	"go.uber.org/zap"
	"nextui-game-manager/models"
	"nextui-game-manager/utils"
	"os/exec"
	"qlova.tech/sum"
	"time"
)

type DownloadArtScreen struct {
	Game                 shared.Item
	RomDirectory         shared.RomDirectory
	SearchFilter         string
	PreviousRomDirectory shared.RomDirectory
	DownloadType         sum.Int[shared.ArtDownloadType]
}

func InitDownloadArtScreen(game shared.Item, romDirectory shared.RomDirectory,
	previousRomDirectory shared.RomDirectory,
	searchFilter string, downloadType sum.Int[shared.ArtDownloadType]) DownloadArtScreen {
	return DownloadArtScreen{
		Game:                 game,
		RomDirectory:         romDirectory,
		PreviousRomDirectory: previousRomDirectory,
		SearchFilter:         searchFilter,
		DownloadType:         downloadType,
	}
}

func (da DownloadArtScreen) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.DownloadArt
}

func (da DownloadArtScreen) Draw() (value models.ScreenReturn, exitCode int, e error) {
	logger := common.GetLoggerInstance()

	ctx := context.Background()
	ctxWithCancel, cancel := context.WithCancel(ctx)
	defer cancel()

	args := []string{"--message", "Attempting to download art...", "--timeout", "-1"}
	cmd := exec.CommandContext(ctxWithCancel, "minui-presenter", args...)

	err := cmd.Start()
	if err != nil && cmd.ProcessState.ExitCode() > 6 {
		logger.Fatal("Error with starting miniui-presenter download art message", zap.Error(err))
	}

	time.Sleep(1250 * time.Millisecond)

	artPath := utils.FindArt(da.Game, da.RomDirectory, da.DownloadType)

	cancel()

	if artPath == "" {
		_, err := cui.ShowMessage("Could not find art :(", "3")
		if err != nil {
			return shared.Item{}, 404, err
		}
		logger.Info("Could not find art!")
		return shared.Item{}, 404, nil
	}

	code, _ := cui.ShowMessageWithOptions("　　　　　　　　　　　　　　　　　　　　　　　　　", "0",
		"--background-image", artPath,
		"--confirm-text", "Use",
		"--confirm-show", "true",
		"--action-button", "X",
		"--action-text", "I'll Find My Own",
		"--action-show", "true",
		"--message-alignment", "bottom")

	if code == 2 || code == 4 {
		common.DeleteFile(artPath)
	}
	return shared.Item{
		Path: artPath,
	}, 0, nil

}
