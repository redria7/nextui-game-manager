package utils

import (
	"database/sql"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"go.uber.org/zap"
	"path/filepath"
	"strings"
)

func HasGameTrackerData(romFilename string, romDirectory shared.RomDirectory) bool {
	db, err := openGameTrackerDB()
	if err != nil {
		return false
	}
	defer closeDB(db)

	gameTrackerRomPath := buildGameTrackerPath(romDirectory.Path, romFilename)

	var romID string
	err = db.QueryRow("SELECT id FROM rom WHERE file_path = ?", gameTrackerRomPath).Scan(&romID)
	return err == nil && romID != ""
}

func buildGameTrackerPath(romDirectoryPath, filename string) string {
	fullPath := filepath.Join(romDirectoryPath, filename)
	return strings.ReplaceAll(fullPath, common.RomDirectory+"/", "")
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

func updateGameTrackerForRename(oldFilename, newFilename string, romDirectory shared.RomDirectory, logger *zap.Logger) {
	oldPath := buildGameTrackerPath(romDirectory.Path, oldFilename)
	newPath := buildGameTrackerPath(romDirectory.Path, newFilename+filepath.Ext(oldFilename))

	logger.Debug("Updating game tracker for rename", zap.String("old", oldPath), zap.String("new", newPath))
	MigrateGameTrackerData(newFilename, oldPath, newPath)
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

func deleteGameTrackerData(tx *sql.Tx, romID string) error {
	if _, err := tx.Exec("DELETE FROM play_activity WHERE rom_id = ?", romID); err != nil {
		return err
	}
	_, err := tx.Exec("DELETE FROM rom WHERE id = ?", romID)
	return err
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
