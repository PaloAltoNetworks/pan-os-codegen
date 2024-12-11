package version

import (
	"fmt"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Version represents PAN-OS version split into its components
type Version struct {
	Major, Minor, Patch int
	Hotfix              string
}

// NewVersionFromString creates a new Version value from a given string
func NewVersionFromString(version string) (Version, error) {
	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return Version{}, fmt.Errorf("invalid version string: not enough components")
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return Version{}, fmt.Errorf("invalid version string: major component must be a number")
	}

	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return Version{}, fmt.Errorf("invalid version string: minor component must be a number")
	}

	patchParts := strings.Split(parts[2], "-")

	var patch int
	var hotfix string
	if len(patchParts) == 1 {
		patch, err = strconv.Atoi(parts[2])
		if err != nil {
			return Version{}, fmt.Errorf("invalid version string: patch component must be a number")
		}
	} else {
		if patchParts[1] == "" {
			return Version{}, fmt.Errorf("invalid version string: hotfix part must be set")
		}
		patch, err = strconv.Atoi(patchParts[0])
		if err != nil {
			return Version{}, fmt.Errorf("invalid version string: patch component must be a number")
		}
		hotfix = patchParts[1]
	}

	return Version{
		Major:  major,
		Minor:  minor,
		Patch:  patch,
		Hotfix: hotfix,
	}, nil

}

// UnmarshalYAML implements custom unmarshalling of YAML data into Version
func (v *Version) UnmarshalYAML(node *yaml.Node) error {
	var versionString string
	if err := node.Decode(&versionString); err != nil {
		return fmt.Errorf("failed to unmarshal YAML structure: %w", err)
	}

	decoded, err := NewVersionFromString(versionString)
	if err != nil {
		return err
	}

	*v = decoded

	return nil
}

// String() returns a string representation of the Version value
func (v Version) String() string {
	version := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	if v.Hotfix != "" {
		version = fmt.Sprintf("%s-%s", version, v.Hotfix)
	}

	return version
}

// LesserThan implements lesser than
//
// This function, just like other boolean comparisons ignore any value
// of the Hotfix field.
func (v Version) LesserThan(o Version) bool {
	if v.Major != o.Major {
		return v.Major < o.Major
	}

	if v.Minor != o.Minor {
		return v.Minor < o.Minor
	}

	return v.Patch < o.Patch
}

// EqualTo implements equality comparison
//
// This function, just like other boolean comparisons ignore any value
// of the Hotfix field.
func (v Version) EqualTo(o Version) bool {
	return v.Major == o.Major && v.Minor == o.Minor && v.Patch == o.Patch
}

// GreaterThan implements greater than comparison
//
// This function, just like other boolean comparisons ignore any value
// of the Hotfix field.
func (v Version) GreaterThan(o Version) bool {
	if v.Major != o.Major {
		return v.Major > o.Major
	}

	if v.Minor != o.Minor {
		return v.Minor > o.Minor
	}

	return v.Patch > o.Patch
}

// GreatherThanOrEqualTo implements greater or equal comparison
//
// This function, just like other boolean comparisons ignore any value
// of the Hotfix field.
func (v Version) GreaterThanOrEqualTo(o Version) bool {
	return v.EqualTo(o) || v.GreaterThan(o)
}

// LesserThanOrEqualTo implemets lesser or equal comparison
//
// This function, just like other boolean comparisons ignore any value
// of the Hotfix field.
func (v Version) LesserThanOrEqualTo(o Version) bool {
	return v.EqualTo(o) || v.LesserThan(o)
}

// SupportedVersionRange calculates a range of supported minor versions
//
// Calculate a range of supported minor versions between min (inclusive) and max (exclusive)
// arguments.
func SupportedPatchVersionRange(minV Version, maxV Version) ([]Version, error) {
	if minV.Major != maxV.Major || minV.Minor != maxV.Minor {
		return nil, fmt.Errorf("minimum and maximum versions must have the same major and minor components")
	}

	if minV.Patch >= maxV.Patch {
		return nil, fmt.Errorf("minimum patch version cannot be equal or higher than maximum")
	}

	var versions []Version
	for i := minV.Patch; i < maxV.Patch; i++ {
		v, _ := NewVersionFromString(fmt.Sprintf("%d.%d.%d", minV.Major, minV.Minor, i))
		versions = append(versions, v)
	}

	return versions, nil
}
