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

			// For any given entry in the list of entries to move check if the entry
			// index is at or after pivot point index, which will require movement
			// set to be generated.
			existingEntryIdx := existingIdxMap[entries[i]]
			switch movement {
			case movementBefore:
				if existingEntryIdx >= existingPivotIdx {
					movementRequired = true
					break
				}
			case movementAfter:
				if existingEntryIdx <= existingPivotIdx {
					movementRequired = true
					break
				}
			}

			if i == 0 {
				continue
			}

			// Check if the entries to be moved have the same order in the existing
			// slice, and if not require a movement set to be generated.
			switch movement {
			case movementBefore:
				if existingIdxMap[entries[i-1]] >= existingEntryIdx {
					movementRequired = true
					break loop

				}
			case movementAfter:
				if existingIdxMap[entries[i-1]] <= existingEntryIdx {
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
	}

	actions, err := GenerateMovements(existing, expected, entries)
	return expected, actions, err
}

func (o PositionAfter) Move(entries []Movable, existing []Movable) ([]MoveAction, error) {
	expected, actions, err := processPivotMovement(entries, existing, o.Pivot, o.Directly, movementBefore)
	if err != nil {
		return nil, err
	}

	return OptimizeMovements(existing, expected, actions, o), nil
}

func (o PositionBefore) Move(entries []Movable, existing []Movable) ([]MoveAction, error) {
	expected, actions, err := processPivotMovement(entries, existing, o.Pivot, o.Directly, movementBefore)
	if err != nil {
		return nil, err
	}

	return OptimizeMovements(existing, expected, actions, o), nil
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

func OptimizeMovements(existing []Movable, expected []Movable, actions []MoveAction, position any) []MoveAction {
	simulated := make([]Movable, len(existing))
	copy(simulated, existing)

	simulatedIdxMap := createIdxMapFor(simulated)
	expectedIdxMap := createIdxMapFor(expected)

	optimized := make([]MoveAction, len(actions))

	switch position.(type) {
	case PositionBefore:
		slog.Debug("OptimizeMovements()", "position", position, "type", fmt.Sprintf("%T", position))
		slices.Reverse(actions)
	case PositionAfter:
		slog.Debug("OptimizeMovements()", "position", position, "type", fmt.Sprintf("%T", position))
	default:
		return actions
	}

	optimizedIdx := 0
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
			targetIdx = simulatedIdxMap[action.Destination] - 1
		case ActionWhereAfter:
			targetIdx = simulatedIdxMap[action.Destination] + 1
		}

		slog.Debug("OptimizeMovements()", "action", action, "currentIdx", currentIdx, "targetIdx", targetIdx)
		if targetIdx != currentIdx {
			optimized[optimizedIdx] = action
			optimizedIdx++
			simulatedIdxMap[action.Movable] = targetIdx
			updateSimulatedIdxMap(&simulatedIdxMap, action.Movable, currentIdx, targetIdx)
		}
	}

	return optimized[:optimizedIdx]
}

func GenerateMovements(existing []Movable, expected []Movable, entries []Movable) ([]MoveAction, error) {
	if len(existing) != len(expected) {
		return nil, ErrSlicesNotEqualLength
	}

	commonSequences := LongestCommonSubstring(existing, expected)

	entriesIdxMap := createIdxMapFor(entries)

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

	slog.Debug("GenerateMovements()", "common", common, "commonStartIdx", commonStartIdx, "commonEndIdx", commonEndIdx)
	var previous Movable
	for _, elt := range entries {
		slog.Debug("GenerateMovements()", "elt", elt, "existing", existingIdxMap[elt], "expected", expectedIdxMap[elt])
		// If existing index for the element matches the expected one, skip it over
		if existingIdxMap[elt] == expectedIdxMap[elt] {
			continue
		}

		// Else, if expected index is 0, generate move action to move it to the top
		if expectedIdxMap[elt] == 0 {
			movements = append(movements, MoveAction{
				Movable:     elt,
				Destination: nil,
				Where:       ActionWhereTop,
			})
			previous = elt
		} else if len(common) == 0 {
			// If, after filtering out all elements that cannot be moved, common sequence
			// is empty we need to move everything element by element.
			movements = append(movements, MoveAction{
				Movable:     elt,
				Destination: previous,
				Where:       ActionWhereAfter,
			})
			previous = elt
		} else {
			// Otherwise if there is some common sequence of elements between existing and expected
			if expectedIdxMap[elt] <= commonStartIdx {
				slog.Debug("GenerateMovements() HELP1")
				// And the expected index of the element is lower than start of the common sequence
				if previous == nil {
					previous = common[0]
				}

				// Generate a movement action for the element to move it directly before the first
				// element of the common sequence.
				movements = append(movements, MoveAction{
					Movable:     elt,
					Destination: previous,
					Where:       ActionWhereBefore,
				})
			} else if expectedIdxMap[elt] > commonEndIdx {
				slog.Debug("GenerateMovements() HELP2")
				// If expected index of the element is larger than index of the last element of the common
				// sequence
				if previous == nil {
					previous = common[commonLen-1]
				}
				// Generate a move to move this element directly behind it.
				movements = append(movements, MoveAction{
					Movable:     elt,
					Destination: previous,
					Where:       ActionWhereAfter,
				})
				previous = elt

			} else if expectedIdxMap[elt] > expectedIdxMap[common[0]] {
				slog.Debug("GenerateMovements() HELP2")
				if previous == nil {
					previous = common[0]
				}
				movements = append(movements, MoveAction{
					Movable:     elt,
					Destination: previous,
					Where:       ActionWhereAfter,
				})
				previous = elt
			}
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

	actions, err := GenerateMovements(existing, expected, entries)
	if err != nil {
		return nil, err
	}

	return OptimizeMovements(existing, expected, actions, o), nil
}

func (o PositionBottom) Move(entries []Movable, existing []Movable) ([]MoveAction, error) {
	entriesIdxMap := createIdxMapFor(entries)

	filtered := removeEntriesFromExisting(existing, func(entry Movable) bool {
		_, ok := entriesIdxMap[entry]
		return ok
	})

	expected := append(filtered, entries...)

	actions, err := GenerateMovements(existing, expected, entries)
	if err != nil {
		return nil, err
	}
	return OptimizeMovements(existing, expected, actions, o), nil
}

func MoveGroup(position Position, entries []Movable, existing []Movable) ([]MoveAction, error) {
	return position.Move(entries, existing)
}
