package dirinfo

import (
	"testing"
)

var (
	motherDirPath = "./tests/motherdir"
)

func TestScanMotherDir(t *testing.T) {
	entries, err := ScanMotherDir(motherDirPath)
	if err != nil {
		t.Fatalf("failed to scan mother dir: %v", err)
	}
	expectedEntries := []*Entry{
		{
			Type:       FileEntry,
			MotherPath: motherDirPath,
			FileList: []*File{
				{
					RelPathToMother: "",
					Name:            "entry1.txt",
					Ext:             ".txt",
					BytesNum:        0,
				},
			},
		},
		{
			Type:       DirEntry,
			MyDirPath:  "entry2",
			MotherPath: motherDirPath,
			FileList: []*File{
				{
					RelPathToMother: "entry2/file21.txt",
					Name:            "file21.txt",
					Ext:             ".txt",
					BytesNum:        0,
				},
				{
					RelPathToMother: "entry2/sub/file22.txt",
					Name:            "file22.txt",
					Ext:             ".txt",
					BytesNum:        file22BytesNum,
				},
			},
		},
	}
	testCheckEntiesSame(t, expectedEntries, entries)
}

func testCheckEntiesSame(t testing.TB, expectedEntries, entries []*Entry) {
	t.Helper()
	if len(entries) != len(expectedEntries) {
		t.Fatalf("entries length not match: %d != %d", len(entries), len(expectedEntries))
	}
	for i := 0; i < len(entries); i++ {
		testCheckEntrySame(t, entries[i], expectedEntries[i])
	}
}

func testCheckEntrySame(t testing.TB, expectEntry, entry *Entry) {
	t.Helper()
	if expectEntry.Type != entry.Type {
		t.Fatalf("entry type not match: %d != %d", expectEntry.Type, entry.Type)
	}
	if expectEntry.MotherPath != entry.MotherPath {
		t.Fatalf("entry mother path not match: %s != %s", expectEntry.MotherPath, entry.MotherPath)
	}
	if expectEntry.MyDirPath != entry.MyDirPath {
		t.Fatalf("entry my dir path not match: %s != %s", expectEntry.MyDirPath, entry.MyDirPath)
	}
	if len(expectEntry.FileList) != len(entry.FileList) {
		t.Fatalf("entry file list length not match: %d != %d", len(expectEntry.FileList), len(entry.FileList))
	}
	for i := 0; i < len(expectEntry.FileList); i++ {
		testCheckFileSame(t, expectEntry.FileList[i], entry.FileList[i])
	}
}

func testCheckFileSame(t testing.TB, expectFile, file *File) {
	t.Helper()
	if expectFile.RelPathToMother != file.RelPathToMother {
		t.Fatalf("file rel path to mother not match: %s != %s", expectFile.RelPathToMother, file.RelPathToMother)
	}
	if expectFile.Name != file.Name {
		t.Fatalf("file name not match: %s != %s", expectFile.Name, file.Name)
	}
	if expectFile.Ext != file.Ext {
		t.Fatalf("file ext not match: %s != %s", expectFile.Ext, file.Ext)
	}
	if expectFile.BytesNum != file.BytesNum {
		t.Fatalf("file bytes num not match: %d != %d", expectFile.BytesNum, file.BytesNum)
	}
}
