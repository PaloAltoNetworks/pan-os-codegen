package version

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Version is the version number struct.
type Version struct {
	Major, Minor, Patch int
	Hotfix              string
}

// Gte tests if this version number is greater than or equal to the argument.
func (v Version) Gte(o Version) bool {
	if v.Major != o.Major {
		return v.Major > o.Major
	}

	if v.Minor != o.Minor {
		return v.Minor > o.Minor
	}

	return v.Patch >= o.Patch
}

// String returns the version number as a string.
func (v Version) String() string {
	if v.Hotfix == "" {
		return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	} else {
		return fmt.Sprintf("%d.%d.%d-%s", v.Major, v.Minor, v.Patch, v.Hotfix)
	}
}

// New returns a version number from the given string.
func New(version string) (Version, error) {
	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return Version{}, errors.New("invalid version")
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return Version{}, fmt.Errorf("major %s is not a number: %w", parts[0], err)
	}

	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return Version{}, fmt.Errorf("minor %s is not a number: %w", parts[0], err)
	}

	patchWithHotfix := strings.Split(parts[2], "-")

	var hotfix string
	if len(patchWithHotfix) == 1 {
		hotfix = ""
	} else if len(patchWithHotfix) == 2 {
		hotfix = patchWithHotfix[1]
	} else {
		return Version{}, fmt.Errorf("patch %s is not formatted as expected", parts[2])
	}

	patch, err := strconv.Atoi(patchWithHotfix[0])
	if err != nil {
		return Version{}, fmt.Errorf("patch %s is not a number: %w", parts[2], err)
	}

	return Version{major, minor, patch, hotfix}, nil
}
