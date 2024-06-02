package dirinfo

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// File represents a file in the motherDir
type File struct {
	RelPathToMother string // Relative paht	to motherDir
	Name            string // Filename
	Ext             string // Extension
	PureName        string // Filename without extension
	BytesNum        int64  // Total Bytes
}

// EntryType represents the type of entry
type EntryType int

const (
	FileEntry EntryType = iota // File Type
	DirEntry                   // Dir Type
)

// Entry represents a file or a dir in the motherDir, as the smallest unit to process
type Entry struct {
	Type       EntryType // Type of entry
	MyDirPath  string    // Dir path
	MotherPath string    // Mother dir path
	FileList   []*File   // File list, if Type is FileEntry, only one element in the list, or recusive files in the dir
}

func (e *Entry) Name() string {
	switch e.Type {
	case FileEntry:
		if len(e.FileList) != 1 || e.FileList[0] == nil {
			return ""
		}
		return e.FileList[0].Name
	case DirEntry:
		return e.MyDirPath
	default:
		return ""
	}
}

func ScanMotherDir(motherPath string) ([]*Entry, error) {
	// check motherDir is dir
	motherDir, err := os.Open(motherPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open motherDir: %v", err)
	}
	motherDirStat, err := motherDir.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat motherDir: %v", err)
	}
	if !motherDirStat.IsDir() {
		return nil, fmt.Errorf("motherDir is not a directory")
	}
	// get all subs in motherDir
	subs, err := os.ReadDir(motherPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read motherDir: %v", err)
	}
	var entries []*Entry
	for _, sub := range subs {
		var entry *Entry
		if sub.IsDir() {
			entry, err = dirEntry(sub, motherPath)
		} else {
			entry, err = fileEntry(sub, motherPath)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to scan entryDir: %v", err)
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func fileEntry(sub fs.DirEntry, motherPath string) (*Entry, error) {
	if sub.IsDir() {
		panic("sub is not a file")
	}
	info, err := sub.Info()
	if err != nil {
		return nil, fmt.Errorf("failed to get sub info: %v", err)
	}
	e := &Entry{
		Type:       FileEntry,
		MotherPath: motherPath,
		FileList: []*File{
			{
				RelPathToMother: sub.Name(),
				Name:            sub.Name(),
				Ext:             filepath.Ext(sub.Name()),
				PureName:        sub.Name()[:len(sub.Name())-len(filepath.Ext(sub.Name()))],
				BytesNum:        info.Size(),
			},
		},
	}
	return e, nil
}

func dirEntry(sub fs.DirEntry, motherPath string) (*Entry, error) {
	if !sub.IsDir() {
		panic("sub is not a dir")
	}
	e := &Entry{
		Type:       DirEntry,
		MyDirPath:  sub.Name(),
		MotherPath: motherPath,
	}
	filepath.WalkDir(filepath.Join(motherPath, sub.Name()), func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("failed to walk dir: %v", err)
		}
		if d.IsDir() {
			return nil
		}
		if path == sub.Name() {
			return nil
		}
		relPathToMother, err := filepath.Rel(motherPath, path)
		if err != nil {
			return fmt.Errorf("failed to get rel path to mother: %v", err)
		}
		info, err := d.Info()
		if err != nil {
			return fmt.Errorf("failed to get sub info: %v", err)
		}
		e.FileList = append(e.FileList, &File{
			RelPathToMother: filepath.ToSlash(relPathToMother),
			Name:            d.Name(),
			Ext:             filepath.Ext(d.Name()),
			PureName:        d.Name()[:len(d.Name())-len(filepath.Ext(d.Name()))],
			BytesNum:        info.Size(),
		})
		return nil
	})
	return e, nil
}
