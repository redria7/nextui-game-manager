package ui

import (
	"bytes"
	"errors"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	"go.uber.org/zap"
	"nextui-game-manager/models"
	"os"
	"os/exec"
	"qlova.tech/sum"
	"strings"
)

type RenameCollectionScreen struct {
	Collection models.Collection
}

func InitRenameCollectionScreen(collection models.Collection) RenameCollectionScreen {
	return RenameCollectionScreen{
		Collection: collection,
	}
}

func (r RenameCollectionScreen) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.CollectionRename
}

func (r RenameCollectionScreen) Draw() (value models.ScreenReturn, exitCode int, e error) {
	logger := common.GetLoggerInstance()

	args := []string{"--initial-value", r.Collection.DisplayName, "--title", "Rename Collection", "--show-hardware-group"}

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
		return models.WrappedString{Contents: r.Collection.DisplayName}, 2, nil
	}

	outValue := stdoutbuf.String()
	_ = stderrbuf.String()

	newFilename := strings.ReplaceAll(outValue, "\n", "")

	return models.NewWrappedString(newFilename), cmd.ProcessState.ExitCode(), nil
}
