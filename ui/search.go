package ui

import (
	"bytes"
	"errors"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	cui "github.com/UncleJunVIP/nextui-pak-shared-functions/ui"
	"go.uber.org/zap"
	"nextui-game-manager/models"
	"os"
	"os/exec"
	"qlova.tech/sum"
	"strings"
)

type Search struct {
	RomDirectory shared.RomDirectory
}

func InitSearch(romDirectory shared.RomDirectory) Search {
	return Search{
		RomDirectory: romDirectory,
	}
}

func (s Search) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.SearchBox
}

func (s Search) Draw() (value models.ScreenReturn, exitCode int, e error) {
	logger := common.GetLoggerInstance()

	args := []string{"--title", "Search " + s.RomDirectory.DisplayName}

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
		_, _ = cui.ShowMessage("Unable to open keyboard!", "3")
		return models.NewWrappedString(""), cmd.ProcessState.ExitCode(), nil
	}

	outValue := strings.TrimSpace(stdoutbuf.String())
	_ = stderrbuf.String()

	return models.NewWrappedString(outValue), cmd.ProcessState.ExitCode(), nil
}
