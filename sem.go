package sem

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// See http://semver.org/spec/v2.0.0.html for semver spec details

var (
	// ErrMultiMeta is returned when attempting to parse a semvar that has
	// more than one + for indicating meta.
	ErrMultiMeta = errors.New("only one +meta is allowed")

	// ErrBadSemVer is returned when attempting to parse a semvar that does
	// not have three normal versions: major.minor.patch preceding an
	// optional prerelease and meta.
	ErrBadSemVer = errors.New("major.minor.patch must be specified")
)

// ParseError reports the location of invalid characters while parsing a semvar
// string with New.
type ParseError struct {
	// T is the section of the semvar where the bad character was found:
	// normal, prerelease, or meta.
	T string

	// R is the character that is the problem.
	R rune

	// P is the numerical position within the section specified by T where
	// the bad character was found. The first character is in position 0.
	P int
}

// Error makes ParseError implement the error interface.
func (err ParseError) Error() string {
	return fmt.Sprintf("bad %s character: '%c', in position %d", err.T, err.R, err.P)
}

// Version contains all of the information to specify a single semvar.
type Version struct {
	// Normal contains the three normal versions: major, minor and patch.
	Normal [3]int

	// Prerelease contains all prerelease version information. For example
	// it could contain the specific rc version [rc, 3].
	Prerelease []string

	// Meta contains the raw string of all meta tags for this version. Meta
	// is not significant for version precedence comparison but could be
	// useful for search or linking.
	Meta string
}

// New creates a new Version from a semvar string and reports any errors
// encountered.
func New(s string) (*Version, error) {
	v := &Version{}

	// extract metadata
	meta := strings.Split(s, "+")
	switch len(meta) {
	case 1: // no meta
	case 2:
		for i, r := range meta[1] {
			if (r < '0' || r > '9') &&
				(r < 'a' || r > 'z') &&
				(r < 'A' || r > 'Z') &&
				r != '-' &&
				r != '.' {
				return nil, ParseError{"meta", r, i}
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
	offset := 0
	for i, versionStr := range normal {
		for j, r := range versionStr {
			if r < '0' || r > '9' {
				return nil, ParseError{"normal", r, offset + j}
			}
		}
		version, err := strconv.Atoi(versionStr)
		if err != nil {
			return nil, err
		}
		v.Normal[i] = version
		offset += len(versionStr) + 1
	}

	// get prerelease version
	pre := strings.Join(longVersion[1:], "-")
	if pre == "" {
		return v, nil
	}
	for i, r := range pre {
		if (r < '0' || r > '9') &&
			(r < 'a' || r > 'z') &&
			(r < 'A' || r > 'Z') &&
			r != '-' &&
			r != '.' {
			return nil, ParseError{"prerelease", r, i}
		}
	}
	v.Prerelease = strings.Split(pre, ".")
	return v, nil
}

// String converts a Version back into a semvar string.
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

// IsAtLeast reports whether the version v has equal or greater precedence than
// the minimum required version specified by min. Versions with greater
// precedence are newer. If a.IsAtLeast(b) that means that version a either has
// the same semantics as version b or version a has newer semantics than
// version b.
func (v *Version) IsAtLeast(min *Version) bool {
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
