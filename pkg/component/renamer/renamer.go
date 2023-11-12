package renamer

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

type Path []string

type RenameRecord struct {
	Old, New Path
}

type Renamer interface {
	Rename(records []RenameRecord) error
}

type FileRenamer struct{}

var _ Renamer = (*FileRenamer)(nil)

func (renamer *FileRenamer) Rename(records []RenameRecord) error {
	for _, v := range records {
		err := renamer.renameOne(v.Old, v.New)
		if err != nil {
			return err
		}
	}
	return nil
}

func (*FileRenamer) renameOne(old, new Path) error {
	log.Printf("renamer old:\n")
	for _, v := range old {
		log.Printf("**** %s\n", v)
	}
	log.Printf("renamer new:\n")
	for _, v := range new {
		log.Printf("**** %s\n", v)
	}

	if len(new) == 0 {
		return fmt.Errorf("new empty")
	}
	targetDir := new[0]
	stat, err := os.Stat(targetDir)
	if err != nil {
		return err
	}
	statMode := stat.Mode()
	targetPath := filepath.Join(new[:len(new)-1]...)
	err = os.MkdirAll(targetPath, statMode)
	if err != nil {
		return err
	}
	err = os.Rename(filepath.Join(old...), filepath.Join(new...))
	if err != nil {
		return err
	}
	return nil
}
