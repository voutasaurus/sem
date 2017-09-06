package sem

import (
	"reflect"
	"testing"
)

func TestNew(t *testing.T) {
	tests := map[string]struct {
		version     string
		wantVersion *Version
		wantError   error
	}{

		"empty": {
			version:   "",
			wantError: ErrBadSemVer,
		},

		"multimeta": {
			version:   "1.1.1-beta+one+two",
			wantError: ErrMultiMeta,
		},

		"bad normal character": {
			version:   "a.b.c",
			wantError: CharacterError{"normal", 'a'},
		},

		"bad prerelease character": {
			version:   "1.0.0-be$tversion",
			wantError: CharacterError{"prerelease", '$'},
		},

		"bad meta character": {
			version:   "1.0.0+blah$",
			wantError: CharacterError{"meta", '$'},
		},

		"simple": {
			version: "1.0.0",
			wantVersion: &Version{
				Normal: [3]int{1, 0, 0},
			},
		},

		"prerelease": {
			version: "1.0.0-beta.1.1",
			wantVersion: &Version{
				Normal:     [3]int{1, 0, 0},
				Prerelease: []string{"beta", "1", "1"},
			},
		},

		"complex": {
			version: "1.0.0-beta.1.1+meta",
			wantVersion: &Version{
				Normal:     [3]int{1, 0, 0},
				Prerelease: []string{"beta", "1", "1"},
				Meta:       "meta",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			gotVersion, gotError := New(tt.version)
			if tt.wantError != gotError {
				t.Errorf("want error: %v, got error: %v", tt.wantError, gotError)
			}
			if !reflect.DeepEqual(tt.wantVersion, gotVersion) {
				t.Errorf("want version: %+v, got version: %+v", tt.wantVersion, gotVersion)
			}
		})
	}
}

func TestAtLeast(t *testing.T) {
	tests := map[string]struct {
		v    *Version
		min  *Version
		want bool
	}{

		"basic": {
			v:    &Version{Normal: [3]int{1, 0, 0}},
			min:  &Version{Normal: [3]int{1, 0, 0}},
			want: true,
		},

		"one more": {
			v:    &Version{Normal: [3]int{1, 0, 1}},
			min:  &Version{Normal: [3]int{1, 0, 0}},
			want: true,
		},

		"one less": {
			v:    &Version{Normal: [3]int{1, 0, 0}},
			min:  &Version{Normal: [3]int{1, 0, 1}},
			want: false,
		},

		"prerelease": {
			v: &Version{
				Normal:     [3]int{1, 0, 0},
				Prerelease: []string{"beta"},
			},
			min:  &Version{Normal: [3]int{1, 0, 0}},
			want: false,
		},

		"prerelease as min": {
			v: &Version{Normal: [3]int{1, 0, 0}},
			min: &Version{
				Normal:     [3]int{1, 0, 0},
				Prerelease: []string{"beta"},
			},
			want: true,
		},

		"prerelease same": {
			v: &Version{
				Normal:     [3]int{1, 0, 0},
				Prerelease: []string{"beta"},
			},
			min: &Version{
				Normal:     [3]int{1, 0, 0},
				Prerelease: []string{"beta"},
			},
			want: true,
		},

		"prerelease diff": {
			v: &Version{
				Normal:     [3]int{1, 0, 0},
				Prerelease: []string{"alpha"},
			},
			min: &Version{
				Normal:     [3]int{1, 0, 0},
				Prerelease: []string{"beta"},
			},
			want: false,
		},

		"prerelease bump": {
			v: &Version{
				Normal:     [3]int{1, 0, 0},
				Prerelease: []string{"beta", "2"},
			},
			min: &Version{
				Normal:     [3]int{1, 0, 0},
				Prerelease: []string{"beta"},
			},
			want: true,
		},

		"prerelease bump requirement": {
			v: &Version{
				Normal:     [3]int{1, 0, 0},
				Prerelease: []string{"beta"},
			},
			min: &Version{
				Normal:     [3]int{1, 0, 0},
				Prerelease: []string{"beta", "2"},
			},
			want: false,
		},

		"prerelease int vs string": {
			v: &Version{
				Normal:     [3]int{1, 0, 0},
				Prerelease: []string{"1"},
			},
			min: &Version{
				Normal:     [3]int{1, 0, 0},
				Prerelease: []string{"beta"},
			},
			want: false,
		},

		"prerelease int": {
			v: &Version{
				Normal:     [3]int{1, 0, 0},
				Prerelease: []string{"1", "0"},
			},
			min: &Version{
				Normal:     [3]int{1, 0, 0},
				Prerelease: []string{"1", "1"},
			},
			want: false,
		},

		"prerelease int same": {
			v: &Version{
				Normal:     [3]int{1, 0, 0},
				Prerelease: []string{"1", "1"},
			},
			min: &Version{
				Normal:     [3]int{1, 0, 0},
				Prerelease: []string{"1", "1"},
			},
			want: true,
		},

		"prerelease int after": {
			v: &Version{
				Normal:     [3]int{1, 0, 0},
				Prerelease: []string{"1", "2"},
			},
			min: &Version{
				Normal:     [3]int{1, 0, 0},
				Prerelease: []string{"1", "1"},
			},
			want: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := tt.v.AtLeast(tt.min)
			if tt.want != got {
				t.Errorf("want AtLeast: %v, got AtLeast: %v", tt.want, got)
			}
		})
	}
}
