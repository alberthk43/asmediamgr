package fileinfocheck

import (
	"asmediamgr/pkg/common"
	"testing"
)

func TestHasOnlyOneFileChecker(t *testing.T) {
	/// arrange
	tests := []struct {
		tname    string
		info     *common.Info
		expectOK bool
	}{
		{
			tname: "one file case",
			info: &common.Info{
				Subs: []common.Single{
					{
						Name: "some file name",
						Ext:  ".mkv",
					},
				},
			},
			expectOK: true,
		},
		{
			tname: "multiple file case",
			info: &common.Info{
				Subs: []common.Single{
					{
						Name: "some file name",
						Ext:  ".mkv",
					},
					{
						Name: "some other file name",
						Ext:  ".mkv",
					},
				},
			},
			expectOK: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.tname, func(t *testing.T) {
			/// action
			err := CheckFileInfo(tt.info, CheckFileInfoHasOnlyOneFile())
			/// assert
			if tt.expectOK && err == nil {
				return
			}
			if !tt.expectOK && err != nil {
				return
			}
			t.Fatalf("expectOK:%t, but err:%v\n", tt.expectOK, err)
		})
	}
}

func TestHasOnlyOneMediaFil(t *testing.T) {
	/// arrange
	tests := []struct {
		tname    string
		info     *common.Info
		expectOK bool
	}{
		{
			tname: "one media file case",
			info: &common.Info{
				Subs: []common.Single{
					{
						Name: "some file name",
						Ext:  ".mkv",
					},
				},
			},
			expectOK: true,
		},
		{
			tname: "one not media file case",
			info: &common.Info{
				Subs: []common.Single{
					{
						Name: "some file name",
						Ext:  ".txt",
					},
				},
			},
			expectOK: false,
		},
		{
			tname: "multiple media file case",
			info: &common.Info{
				Subs: []common.Single{
					{
						Name: "some media file",
						Ext:  ".mkv",
					},
					{
						Name: "another media file",
						Ext:  ".mkv",
					},
				},
			},
			expectOK: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.tname, func(t *testing.T) {
			err := CheckFileInfo(tt.info, CheckFileInfoHasOnlyOneMediaFile())
			if tt.expectOK && err == nil {
				return
			}
			if !tt.expectOK && err != nil {
				return
			}
			t.Fatalf("expect:%t, but err:%v\n", tt.expectOK, err)
		})
	}
}

func TestHasOnlyOneFileBiggerThan(t *testing.T) {
	/// arrange
	tests := []struct {
		tname    string
		info     *common.Info
		expectOk bool
	}{
		{
			tname: "one file bigger than 50mB case",
			info: &common.Info{
				Subs: []common.Single{
					{
						Name: "some file name",
						Ext:  ".mkv",
						Size: 1024 * 1024 * 1024,
					},
				},
			},
			expectOk: true,
		},
		{
			tname: "one file smaller than 50mB case",
			info: &common.Info{
				Subs: []common.Single{
					{
						Name: "some file name",
						Ext:  ".mkv",
						Size: 49 * 1024 * 1024,
					},
				},
			},
			expectOk: false,
		},
		{
			tname: "multiple file smaller than 50mB case",
			info: &common.Info{
				Subs: []common.Single{
					{
						Name: "some file name",
						Ext:  ".mkv",
						Size: 49 * 1024 * 1024,
					},
					{
						Name: "another some file name",
						Ext:  ".mkv",
						Size: 49 * 1024 * 1024,
					},
				},
			},
			expectOk: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.tname, func(t *testing.T) {
			err := CheckFileInfo(tt.info, CheckFileInfoHasOnlyOneFileSizeGreaterThan(50*1024*1024))
			if tt.expectOk && err == nil {
				return
			}
			if !tt.expectOk && err != nil {
				return
			}
			t.Fatalf("expect:%t, but err:%v\n", tt.expectOk, err)
		})
	}
}
