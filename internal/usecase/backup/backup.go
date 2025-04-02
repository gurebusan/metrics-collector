package backup

import (
	"fmt"
	"os"
	"path/filepath"
)

type ServerBackupMnger interface {
	SaveBkpToFile(path string) error
	LoadBkpFromFile(path string) error
}

type BackupUsecase struct {
	mnger ServerBackupMnger
}

func NewBackupUsecase(mnger ServerBackupMnger) *BackupUsecase {
	return &BackupUsecase{mnger: mnger}
}

func (b *BackupUsecase) SaveBackup(relativePath string) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	path := filepath.Join(wd, relativePath)

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	if err = b.mnger.SaveBkpToFile(path); err != nil {
		return err
	}
	return nil
}

func (b *BackupUsecase) LoadBackup(relativePath string) error {

	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	path := filepath.Join(wd, relativePath)

	if err = b.mnger.LoadBkpFromFile(path); err != nil {
		return err
	}
	return nil
}
