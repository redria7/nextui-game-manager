package ui

import (
	"fmt"
	"github.com/UncleJunVIP/gabagool/pkg/gabagool"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"go.uber.org/zap"
	"nextui-game-manager/models"
	"nextui-game-manager/utils"
	"qlova.tech/sum"
	"time"
)

type ArchiveOptionsScreen struct {
	Archive   shared.RomDirectory
}

func InitArchiveOptionsScreen(archive shared.RomDirectory) ArchiveOptionsScreen {
	return ArchiveOptionsScreen{
		Archive:   archive,
	}
}

func (aos ArchiveOptionsScreen) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.ArchiveOptions
}

func (aos ArchiveOptionsScreen) Draw() (screenReturn interface{}, exitCode int, e error) {
	logger := common.GetLoggerInstance()

	var actions []gabagool.MenuItem
	for _, action := range models.ArchiveActionKeys {
		actions = append(actions, gabagool.MenuItem{
			Text:     action,
			Selected: false,
			Focused:  false,
			Metadata: action,
		})
	}

	options := gabagool.DefaultListOptions(fmt.Sprintf("%s Options", aos.Archive.DisplayName), actions)
	options.EnableAction = true
	options.FooterHelpItems = []gabagool.FooterHelpItem{
		{ButtonName: "B", HelpText: "Back"},
		{ButtonName: "A", HelpText: "Select"},
	}

	result, err := gabagool.List(options)

	if err != nil {
		return nil, -1, err
	}

	if result.IsSome() && !result.Unwrap().ActionTriggered && result.Unwrap().SelectedIndex != -1 {
		action := models.ActionMap[result.Unwrap().SelectedItem.Metadata.(string)]

		switch action {
		case models.Actions.ArchiveRename:
			oldArchive := utils.CleanArchiveName(aos.Archive.DisplayName)
			res, err := gabagool.Keyboard(oldArchive)

			if err != nil {
				return nil, 1, err
			}

			if res.IsSome() {
				newArchive := res.Unwrap()
				if newArchive != oldArchive {
					newArchive = utils.PrepArchiveName(newArchive)
					newArchivePath := utils.GetArchiveRoot(newArchive)

					err := utils.MoveFile(aos.Archive.Path, newArchivePath)

					if err != nil {
						logger.Error("Failed to rename archive", zap.Error(err))
						utils.ShowTimedMessage("Failed to rename archive", time.Second * 2)
						return nil, 1, err
					}

					archiveDirectory := shared.RomDirectory{
						DisplayName: newArchive,
						Path: 		 newArchivePath,
					}

					return archiveDirectory, 4, nil
				}
			}

			return nil, 4, nil

		case models.Actions.ArchiveDelete:
			res, _ := gabagool.ConfirmationMessage(fmt.Sprintf("Are you sure you want to delete the archive\n%s?", aos.Archive.DisplayName), []gabagool.FooterHelpItem{
				{ButtonName: "B", HelpText: "Cancel"},
				{ButtonName: "X", HelpText: "Delete"},
			}, gabagool.MessageOptions{
				ImagePath:     "",
				ConfirmButton: gabagool.ButtonX,
			})

			if res.IsSome() && !res.Unwrap().Cancelled {
				res, err := utils.DeleteArchive(aos.Archive)

				if err != nil {
					logger.Error("Failed to delete archive", zap.Error(err))
					utils.ShowTimedMessage("Failed to delete archive", time.Second * 2)
					return nil, 1, err
				}

				if res != "" {
					utils.ShowTimedMessage(fmt.Sprintf("Cannot delete while file exists in archive\n%s", res), time.Second * 2)
					return aos.Archive, 2, nil
				}

				return nil, 0, nil
			}

		}

		return aos.Archive, 2, nil
	}

	return aos.Archive, 2, nil
}
