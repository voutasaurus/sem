package sem

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// See http://semver.org/spec/v2.0.0.html for semver spec details

var (
	ErrMultiMeta = errors.New("only one +meta is allowed")
	ErrBadSemVer = errors.New("major.minor.patch must be specified")
)

type CharacterError struct {
	T string
	R rune
}

func (err CharacterError) Error() string {
	return fmt.Sprintf("bad %s character: %c", err.T, err.R)
}

type Version struct {
	Normal     [3]int
	Prerelease []string
	Meta       string
}

func New(s string) (*Version, error) {
	v := &Version{}

	// extract metadata
	meta := strings.Split(s, "+")
	switch len(meta) {
	case 1: // no meta
	case 2:
		for _, r := range meta[1] {
			if (r < '0' || r > '9') &&
				(r < 'a' || r > 'z') &&
				(r < 'A' || r > 'Z') &&
				r != '-' &&
				r != '.' {
				return nil, CharacterError{"meta", r}
			}
		}
		v.Meta = meta[1]
	default:
		return nil, ErrMultiMeta
	}

	// get normal version
	longVersion := strings.Split(meta[0], "-")
	normal := strings.Split(longVersion[0], ".")
	if len(normal) != 3 {
		return nil, ErrBadSemVer
	}
	for i, versionStr := range normal {
		for _, r := range versionStr {
			if r < '0' || r > '9' {
				return nil, CharacterError{"normal", r}
			}
		}
		version, err := strconv.Atoi(versionStr)
		if err != nil {
			return nil, err
		}
		v.Normal[i] = version
	}

	// get prerelease version
	pre := strings.Join(longVersion[1:], "-")
	if pre == "" {
		return v, nil
	}
	for _, r := range pre {
		if (r < '0' || r > '9') &&
			(r < 'a' || r > 'z') &&
			(r < 'A' || r > 'Z') &&
			r != '-' &&
			r != '.' {
			return nil, CharacterError{"prerelease", r}
		}
	}
	v.Prerelease = strings.Split(pre, ".")
	return v, nil
}

func (v *Version) String() string {
	s := fmt.Sprintf("%d.%d.%d", v.Normal[0], v.Normal[1], v.Normal[2])
	pre := strings.Join(v.Prerelease, ".")
	if pre != "" {
		s += "-" + pre
	}
	if v.Meta != "" {
		s += "+" + v.Meta
	}
	return s
}

func (v *Version) AtLeast(min *Version) bool {
	const (
		vIsAtLeastMin  = true
		vIsLessThanMin = false
	)

	for i, vNormal := range v.Normal {
		if vNormal < min.Normal[i] {
			return vIsLessThanMin
		}
		if vNormal > min.Normal[i] {
			return vIsAtLeastMin
		}
	}

	// Versions with no prerelease specified are after those with a
	// prerelease specified.
	if len(v.Prerelease) == 0 && len(min.Prerelease) > 0 {
		return vIsAtLeastMin
	}
	if len(v.Prerelease) > 0 && len(min.Prerelease) == 0 {
		return vIsLessThanMin
	}

	for i := range min.Prerelease {
		if len(v.Prerelease) <= i {
			// v.pre is shorter than min.pre. Longer is higher
			// precendence when all previous identifiers are equal.
			return vIsLessThanMin
		}

		// try to compare current component as int
		minInt := true
		minV, err := strconv.Atoi(min.Prerelease[i])
		if err != nil {
			minInt = false
		}
		vInt := true
		vV, err := strconv.Atoi(v.Prerelease[i])
		if err != nil {
			vInt = false
		}
		if minInt && vInt {
			if vV > minV {
				return vIsAtLeastMin
			}
			if vV < minV {
				return vIsLessThanMin
			}
			continue // equal, check next component
		}

		// strings have higher precedence than ints
		if minInt { // v.pre[i] is string but min.pre[i] is int
			return vIsAtLeastMin
		}
		if vInt { // v.pre[i] is int but min.pre[i] is string
			return vIsLessThanMin
		}

		// compare as string
		if v.Prerelease[i] > min.Prerelease[i] {
			return vIsAtLeastMin
		}
		if v.Prerelease[i] < min.Prerelease[i] {
			return vIsLessThanMin
		}
	}
	// either min.pre is a prefix of v.pre or they are the same
	return vIsAtLeastMin
}
