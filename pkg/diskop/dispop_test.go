package diskop

import "testing"

func TestMakeFilenameWindowsFriendly(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{name: "Perché di Jennifer?", want: "Perché di Jennifer"},
		{name: "Perché di ?Jennifer", want: "Perché di  Jennifer"},
		{name: `Slash\Haha`, want: "Slash Haha"},
		{name: `Col:`, want: "Col"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := replaceSpecialChars(tt.name); got != tt.want {
				t.Errorf("MakeFilenameWindowsFriendly() \nreal %v, \nwant %v", got, tt.want)
			}
		})
	}
}
