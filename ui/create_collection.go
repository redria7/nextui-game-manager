package ui

import (
	"bytes"
	"errors"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/ui"
	"go.uber.org/zap"
	"nextui-game-manager/models"
	"nextui-game-manager/utils"
	"os"
	"os/exec"
	"path/filepath"
	"qlova.tech/sum"
	"strings"
)

type CreateCollectionScreen struct {
	Game                 shared.Item
	RomDirectory         shared.RomDirectory
	PreviousRomDirectory shared.RomDirectory
	SearchFilter         string
}

func InitCreateCollectionScreen(game shared.Item, romDirectory shared.RomDirectory,
	previousRomDirectory shared.RomDirectory, searchFilter string) CreateCollectionScreen {
	return CreateCollectionScreen{
		Game:                 game,
		RomDirectory:         romDirectory,
		PreviousRomDirectory: previousRomDirectory,
		SearchFilter:         searchFilter,
	}
}

func (c CreateCollectionScreen) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.CollectionCreate
}

func (c CreateCollectionScreen) Draw() (collection models.ScreenReturn, exitCode int, e error) {
	logger := common.GetLoggerInstance()

	args := []string{"--title", "Create New Collection", "--show-hardware-group"}

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
		return models.WrappedString{}, 1, err
	}

	if cmd.ProcessState.ExitCode() == 2 {
		return models.WrappedString{}, 2, nil
	}

	outValue := stdoutbuf.String()
	_ = stderrbuf.String()

	newCollectionName := strings.ReplaceAll(outValue, "\n", "")

	_, err = utils.AddCollectionGame(models.Collection{
		DisplayName:    newCollectionName,
		CollectionFile: filepath.Join(common.CollectionDirectory, newCollectionName+".txt"),
	}, c.Game)
	if err != nil {
		_, _ = ui.ShowMessage("Failed to add game to collection!", "3")
		return models.WrappedString{}, 1, err
	}

	return models.NewWrappedString(newCollectionName), cmd.ProcessState.ExitCode(), nil
}
