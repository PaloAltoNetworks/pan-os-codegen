package movement

import (
	"errors"
	"fmt"
	"log/slog"
	"slices"
)

var _ = slog.LevelDebug

type ActionWhereType string

const (
	ActionWhereTop    ActionWhereType = "top"
	ActionWhereBottom ActionWhereType = "bottom"
	ActionWhereBefore ActionWhereType = "before"
	ActionWhereAfter  ActionWhereType = "after"
)

type entryPositionType int

const (
	entryPositionBefore entryPositionType = iota
	entryPositionAfter
)

type Movable interface {
	EntryName() string
}

type MoveAction struct {
	Movable     Movable
	Where       ActionWhereType
	Destination Movable
}

type Position interface {
	Move(entries []Movable, existing []Movable) ([]MoveAction, error)
}

type PositionTop struct{}

type PositionBottom struct{}

type PositionBefore struct {
	Directly bool
	Pivot    Movable
}

type PositionAfter struct {
	Directly bool
	Pivot    Movable
}

func createIdxMapFor(entries []Movable) map[Movable]int {
	entriesIdxMap := make(map[Movable]int, len(entries))
	for idx, elt := range entries {
		entriesIdxMap[elt] = idx
	}
	return entriesIdxMap
}

func removeEntriesFromExisting(entries []Movable, filterFn func(entry Movable) bool) []Movable {
	entryNames := make(map[string]bool, len(entries))
	for _, elt := range entries {
		entryNames[elt.EntryName()] = true
	}

	filtered := make([]Movable, len(entries))
	copy(filtered, entries)

	filtered = slices.DeleteFunc(filtered, filterFn)

	return filtered
}

func findPivotIdx(entries []Movable, pivot Movable) int {
	return slices.IndexFunc(entries, func(entry Movable) bool {
		if entry.EntryName() == pivot.EntryName() {
			return true
		}

		return false
	})

}

type movementType int

const (
	movementBefore movementType = iota
	movementAfter
)

var (
	ErrSlicesNotEqualLength = errors.New("existing and expected slices length mismatch")
	ErrPivotInEntries       = errors.New("pivot element found in the entries slice")
	ErrPivotNotInExisting   = errors.New("pivot element not foudn in the existing slice")
	ErrInvalidMovementPlan  = errors.New("created movement plan is invalid")
)

// PositionBefore and PositionAfter are similar enough that we can generate expected sequences
// for both using the same code and some conditionals based on the given movement.
func processPivotMovement(entries []Movable, existing []Movable, pivot Movable, direct bool, movement movementType) ([]Movable, []MoveAction, error) {
	existingIdxMap := createIdxMapFor(existing)

	entriesPivotIdx := findPivotIdx(entries, pivot)
	if entriesPivotIdx != -1 {
		return nil, nil, ErrPivotInEntries
	}

	existingPivotIdx := findPivotIdx(existing, pivot)
	if existingPivotIdx == -1 {
		return nil, nil, ErrPivotNotInExisting
	}

	if !direct {
		movementRequired := false
		entriesLen := len(entries)
	loop:
		for i := 0; i < entriesLen; i++ {
			existingEntryIdx := existingIdxMap[entries[i]]
			slog.Debug("generate()", "i", i, "len(entries)", len(entries), "entry", entries[i], "existingEntryIdx", existingEntryIdx, "existingPivotIdx", existingPivotIdx)
			// For any given entry in the list of entries to move check if the entry
			// index is at or after pivot point index, which will require movement
			// set to be generated.

			// Then check if the entries to be moved have the same order in the existing
			// slice, and if not require a movement set to be generated.
			switch movement {
			case movementBefore:
				if existingEntryIdx >= existingPivotIdx {
					movementRequired = true
					break
				}

				if i == 0 {
					continue
				}

				if existingIdxMap[entries[i-1]] >= existingEntryIdx {
					movementRequired = true
					break loop

				}
			case movementAfter:
				if existingEntryIdx <= existingPivotIdx {
					movementRequired = true
					break
				}

				if i == len(entries)-1 {
					continue
				}

				if existingIdxMap[entries[i+1]] < existingEntryIdx {
					movementRequired = true
					break loop

				}

			}
		}

		if !movementRequired {
			return nil, nil, nil
		}
	}

	expected := make([]Movable, len(existing))

	entriesIdxMap := createIdxMapFor(entries)

	filtered := removeEntriesFromExisting(existing, func(entry Movable) bool {
		_, ok := entriesIdxMap[entry]
		return ok
	})

	filteredPivotIdx := findPivotIdx(filtered, pivot)

	switch movement {
	case movementBefore:
		expectedIdx := 0
		for ; expectedIdx < filteredPivotIdx; expectedIdx++ {
			expected[expectedIdx] = filtered[expectedIdx]
		}

		for _, elt := range entries {
			expected[expectedIdx] = elt
			expectedIdx++
		}

		expected[expectedIdx] = pivot
		expectedIdx++

		filteredLen := len(filtered)
		for i := filteredPivotIdx + 1; i < filteredLen; i++ {
			expected[expectedIdx] = filtered[i]
			expectedIdx++
		}
	case movementAfter:
		expectedIdx := 0
		for ; expectedIdx < len(filtered); expectedIdx++ {
			expected[expectedIdx] = filtered[expectedIdx]
		}

		for _, elt := range entries {
			expected[expectedIdx] = elt
			expectedIdx++
		}

		filteredLen := len(filtered)
		for i := filteredPivotIdx + 1; i < filteredLen-1; i++ {
			expected[expectedIdx] = filtered[i]
			expectedIdx++
		}
	}

	actions, err := GenerateMovements(existing, expected, entries, movement)
	return expected, actions, err
}

func (o PositionAfter) Move(entries []Movable, existing []Movable) ([]MoveAction, error) {
	expected, actions, err := processPivotMovement(entries, existing, o.Pivot, o.Directly, movementAfter)
	if err != nil {
		return nil, err
	}

	return OptimizeMovements(existing, expected, entries, actions, o), nil
}

func (o PositionBefore) Move(entries []Movable, existing []Movable) ([]MoveAction, error) {
	expected, actions, err := processPivotMovement(entries, existing, o.Pivot, o.Directly, movementBefore)
	if err != nil {
		return nil, err
	}

	return OptimizeMovements(existing, expected, entries, actions, o), nil
}

type Entry struct {
	Element  Movable
	Expected int
	Existing int
}

type sequencePosition struct {
	Start int
	End   int
}

func updateSimulatedIdxMap(idxMap *map[Movable]int, moved Movable, startingIdx int, targetIdx int) {
	slog.Debug("updateSimulatedIdxMap", "entries", idxMap)
	for entry, idx := range *idxMap {
		if entry == moved {
			continue
		}

		slog.Debug("updateSimulatedIdxMap", "entry", entry, "idx", idx, "startingIdx", startingIdx, "targetIdx", targetIdx)
		if startingIdx > targetIdx && idx >= targetIdx {
			(*idxMap)[entry] = idx + 1
		} else if startingIdx < targetIdx && idx >= startingIdx && idx <= targetIdx {
			(*idxMap)[entry] = idx - 1
		}
	}
	slog.Debug("updateSimulatedIdxMap", "entries", idxMap)
}

func OptimizeMovements(existing []Movable, expected []Movable, entries []Movable, actions []MoveAction, position any) []MoveAction {
	simulated := make([]Movable, len(existing))
	copy(simulated, existing)

	simulatedIdxMap := createIdxMapFor(simulated)
	expectedIdxMap := createIdxMapFor(expected)

	var optimized []MoveAction

	switch position.(type) {
	case PositionBefore:
		slog.Debug("OptimizeMovements()", "position", position, "type", fmt.Sprintf("%T", position))
	case PositionAfter:
		slog.Debug("OptimizeMovements()", "position", position, "type", fmt.Sprintf("%T", position))
	default:
		return actions
	}

	for _, action := range actions {
		currentIdx := simulatedIdxMap[action.Movable]
		if currentIdx == expectedIdxMap[action.Movable] {
			continue
		}

		var targetIdx int
		switch action.Where {
		case ActionWhereTop:
			targetIdx = 0
		case ActionWhereBottom:
			targetIdx = len(simulated) - 1
		case ActionWhereBefore:
			slog.Debug("OptimizeMovements()", "dest", action.Destination, "destIdx", simulatedIdxMap[action.Destination])
			targetIdx = simulatedIdxMap[action.Destination] - 1
		case ActionWhereAfter:
			targetIdx = simulatedIdxMap[action.Destination] + 1
		}

		slog.Debug("OptimizeMovements()", "action", action, "currentIdx", currentIdx, "targetIdx", targetIdx)
		if targetIdx != currentIdx {
			optimized = append(optimized, action)
			simulatedIdxMap[action.Movable] = targetIdx
			updateSimulatedIdxMap(&simulatedIdxMap, action.Movable, currentIdx, targetIdx)
		}
	}

	slog.Debug("OptimizeMovements()", "optimized", optimized)

	return optimized
}

func GenerateMovements(existing []Movable, expected []Movable, entries []Movable, movement movementType) ([]MoveAction, error) {
	if len(existing) != len(expected) {
		return nil, ErrSlicesNotEqualLength
	}

	commonSequences := LongestCommonSubstring(existing, expected)

	entriesIdxMap := createIdxMapFor(entries)

	// LCS returns a list of longest common sequences found between existing and expected
	// slices. We want to find the longest common sequence that doesn't intersect entries
	// given by the user, as entries are moved in relation to the common sequence.
	var common []Movable
	for _, sequence := range commonSequences {
		filtered := removeEntriesFromExisting(sequence, func(elt Movable) bool {
			_, ok := entriesIdxMap[elt]
			return ok
		})

		if len(filtered) > len(common) {
			common = filtered
		}

	}
	commonLen := len(common)

	existingIdxMap := createIdxMapFor(existing)
	expectedIdxMap := createIdxMapFor(expected)

	var movements []MoveAction

	var commonStartIdx, commonEndIdx int
	if commonLen > 0 {
		commonStartIdx = expectedIdxMap[common[0]]
		commonEndIdx = expectedIdxMap[common[commonLen-1]]
	}

	slog.Debug("GenerateMovements()", "expected", expected)
	slog.Debug("GenerateMovements()", "existing", existing)
	slog.Debug("GenerateMovements()", "common", common, "commonStartIdx", commonStartIdx, "commonEndIdx", commonEndIdx)
	var previous Movable
	for _, elt := range entries {
		slog.Debug("GenerateMovements()", "elt", elt, "existing", existingIdxMap[elt], "expected", expectedIdxMap[elt])
		// If existing index for the element matches the expected one, skip it over
		if existingIdxMap[elt] == expectedIdxMap[elt] {
			continue
		}

		if expectedIdxMap[elt] == 0 {
			slog.Debug("HELP1")
			movements = append(movements, MoveAction{
				Movable:     elt,
				Destination: nil,
				Where:       ActionWhereTop,
			})
			previous = elt
		} else if expectedIdxMap[elt] == len(expectedIdxMap) {
			slog.Debug("HELP2")
			movements = append(movements, MoveAction{
				Movable:     elt,
				Destination: nil,
				Where:       ActionWhereBottom,
			})
			previous = elt
		} else if previous != nil {
			slog.Debug("HELP3")

			movements = append(movements, MoveAction{
				Movable:     elt,
				Destination: previous,
				Where:       ActionWhereAfter,
			})
			previous = elt
		} else {
			slog.Debug("HELP4")
			var where ActionWhereType

			switch movement {
			case movementAfter:
				previous = common[commonLen-1]
				where = ActionWhereAfter
			case movementBefore:
				previous = common[0]
				where = ActionWhereBefore
			}

			movements = append(movements, MoveAction{
				Movable:     elt,
				Destination: previous,
				Where:       where,
			})
			previous = elt
		}
	}

	_ = previous

	slog.Debug("GenerateMovements()", "movements", movements)

	return movements, nil
}

func (o PositionTop) Move(entries []Movable, existing []Movable) ([]MoveAction, error) {
	entriesIdxMap := createIdxMapFor(entries)

	filtered := removeEntriesFromExisting(existing, func(entry Movable) bool {
		_, ok := entriesIdxMap[entry]
		return ok
	})

	expected := append(entries, filtered...)

	actions, err := GenerateMovements(existing, expected, entries, movementBefore)
	if err != nil {
		return nil, err
	}

	return OptimizeMovements(existing, expected, entries, actions, o), nil
}

func (o PositionBottom) Move(entries []Movable, existing []Movable) ([]MoveAction, error) {
	entriesIdxMap := createIdxMapFor(entries)

	filtered := removeEntriesFromExisting(existing, func(entry Movable) bool {
		_, ok := entriesIdxMap[entry]
		return ok
	})

	expected := append(filtered, entries...)

	actions, err := GenerateMovements(existing, expected, entries, movementAfter)
	if err != nil {
		return nil, err
	}
	return OptimizeMovements(existing, expected, entries, actions, o), nil
}

func MoveGroup(position Position, entries []Movable, existing []Movable) ([]MoveAction, error) {
	return position.Move(entries, existing)
}

// Debug helper to print generated LCS matrix
func printLCSMatrix(S []Movable, T []Movable, L [][]int) {
	r := len(S)
	n := len(T)

	line := "      "
	for _, elt := range S {
		line += fmt.Sprintf("%s  ", elt.EntryName())
	}
	slog.Debug("LCS", "line", line)

	line = "   "
	for _, elt := range L[0] {
		line += fmt.Sprintf("%d  ", elt)
	}
	slog.Debug("LCS", "line", line)

	for i := 1; i < r+1; i++ {
		line = fmt.Sprintf("%s  ", T[i-1].EntryName())
		for j := 0; j < n+1; j++ {
			line += fmt.Sprintf("%d  ", L[i][j])
		}
	}

}

// LongestCommonSubstring implements dynamic programming variant of the algorithm
//
// See https://en.wikipedia.org/wiki/Longest_common_substring for the details. Our
// implementation is not optimal, as generation of the matrix can be done at the
// same time as finding LCSs, but it's easier to reason about for now.
func LongestCommonSubstring(S []Movable, T []Movable) [][]Movable {
	r := len(S)
	n := len(T)

	L := make([][]int, r+1)
	for idx := range r + 1 {
		L[idx] = make([]int, n+1)
	}

	for i := 1; i < r+1; i++ {
		for j := 1; j < n+1; j++ {
			if S[i-1].EntryName() == T[j-1].EntryName() {
				if i == 1 {
					L[j][i] = 1
				} else if j == 1 {
					L[j][i] = 1
				} else {
					L[j][i] = L[j-1][i-1] + 1
				}
			}
		}
	}

	var results [][]Movable
	var lcsList [][]Movable

	var entry []Movable
	var index int
	for i := r; i > 0; i-- {
		for j := n; j > 0; j-- {
			if S[i-1].EntryName() == T[j-1].EntryName() {
				if L[j][i] >= index {
					if len(entry) > 0 {
						var entries []string
						for _, elt := range entry {
							entries = append(entries, elt.EntryName())
						}

						lcsList = append(lcsList, entry)
					}
					index = L[j][i]
					entry = []Movable{S[i-1]}
				} else if L[j][i] < index {
					index = L[j][i]
					entry = append(entry, S[i-1])
				} else {
					entry = []Movable{}
				}
			}
		}
	}

	if len(entry) > 0 {
		lcsList = append(lcsList, entry)
	}

	lcsLen := len(lcsList)
	for idx := range lcsList {
		elt := lcsList[lcsLen-idx-1]
		if len(elt) > 1 {
			slices.Reverse(elt)
			results = append(results, elt)
		}
	}

	return results
}
