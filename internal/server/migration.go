package server

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"swd-new/internal/model"
	"swd-new/pkg/log"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Migration struct {
	db     *gorm.DB
	logger *log.Logger
}

func NewMigration(conf *viper.Viper, logger *log.Logger) *Migration {
	db, err := gorm.Open(postgres.Open(conf.GetString("data.postgres.dsn")), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	return &Migration{
		db:     db,
		logger: logger,
	}
}

func (m *Migration) SyncSensitiveWordsFromAssets(ctx context.Context) error {
	assetDir := "assets"
	entries, err := os.ReadDir(assetDir)
	if err != nil {
		return fmt.Errorf("read assets dir: %w", err)
	}

	categoryByFile := map[string]model.Word{
		"pornography.txt":    {Type: 1},
		"political.txt":      {Type: 2},
		"violence.txt":       {Type: 4},
		"gambling.txt":       {Type: 8},
		"drugs.txt":          {Type: 16},
		"profanity.txt":      {Type: 32},
		"discrimination.txt": {Type: 64},
		"scam.txt":           {Type: 128},
	}

	words := make([]model.Word, 0, 1024)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		baseWord, ok := categoryByFile[entry.Name()]
		if !ok {
			continue
		}

		filePath := filepath.Join(assetDir, entry.Name())
		file, err := os.Open(filePath)
		if err != nil {
			return fmt.Errorf("open %s: %w", filePath, err)
		}

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			word := strings.TrimSpace(scanner.Text())
			if word == "" {
				continue
			}

			words = append(words, model.Word{
				Word: word,
				Type: baseWord.Type,
			})
		}

		if err := scanner.Err(); err != nil {
			file.Close()
			return fmt.Errorf("scan %s: %w", filePath, err)
		}
		if err := file.Close(); err != nil {
			return fmt.Errorf("close %s: %w", filePath, err)
		}
	}

	if len(words) == 0 {
		m.logger.Info("no sensitive words found in assets")
		return nil
	}

	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.Word{}); err != nil {
			return fmt.Errorf("auto migrate sensitive_words: %w", err)
		}

		if err := tx.Where("1 = 1").Delete(&model.Word{}).Error; err != nil {
			return fmt.Errorf("clear sensitive_words: %w", err)
		}

		if err := tx.CreateInBatches(words, 200).Error; err != nil {
			return fmt.Errorf("insert sensitive words: %w", err)
		}

		m.logger.Info("sensitive words reimported from assets", zap.Int("count", len(words)))
		return nil
	})
}
