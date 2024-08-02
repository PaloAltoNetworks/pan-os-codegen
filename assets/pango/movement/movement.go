package movement

import (
	"errors"
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
	GetExpected(entries []Movable, existing []Movable) ([]Movable, error)
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

var (
	errNoMovements          = errors.New("no movements needed")
	ErrSlicesNotEqualLength = errors.New("existing and expected slices length mismatch")
	ErrPivotInEntries       = errors.New("pivot element found in the entries slice")
	ErrPivotNotInExisting   = errors.New("pivot element not foudn in the existing slice")
	ErrInvalidMovementPlan  = errors.New("created movement plan is invalid")
)

// PositionBefore and PositionAfter are similar enough that we can generate expected sequences
// for both using the same code and some conditionals based on the given movement.
func getPivotMovement(entries []Movable, existing []Movable, pivot Movable, direct bool, movement ActionWhereType) ([]Movable, error) {
	existingIdxMap := createIdxMapFor(existing)

	entriesPivotIdx := findPivotIdx(entries, pivot)
	if entriesPivotIdx != -1 {
		return nil, ErrPivotInEntries
	}

	existingPivotIdx := findPivotIdx(existing, pivot)
	if existingPivotIdx == -1 {
		return nil, ErrPivotNotInExisting
	}

	if !direct {
		movementRequired := false
		entriesLen := len(entries)
	loop:
		for i := 0; i < entriesLen; i++ {
			existingEntryIdx := existingIdxMap[entries[i]]
			// For any given entry in the list of entries to move check if the entry
			// index is at or after pivot point index, which will require movement
			// set to be generated.

			// Then check if the entries to be moved have the same order in the existing
			// slice, and if not require a movement set to be generated.
			switch movement {
			case ActionWhereBefore:
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
			case ActionWhereAfter:
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
			return nil, errNoMovements
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
	case ActionWhereBefore:
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

	case ActionWhereAfter:
		expectedIdx := 0
		for ; expectedIdx < filteredPivotIdx+1; expectedIdx++ {
			expected[expectedIdx] = filtered[expectedIdx]
		}

		if direct {
			for _, elt := range entries {
				expected[expectedIdx] = elt
				expectedIdx++
			}

			filteredLen := len(filtered)
			for i := filteredPivotIdx + 1; i < filteredLen; i++ {
				expected[expectedIdx] = filtered[i]
			}
		} else {
			filteredLen := len(filtered)
			for i := filteredPivotIdx + 1; i < filteredLen; i++ {
				expected[expectedIdx] = filtered[i]
				expectedIdx++
			}

			for _, elt := range entries {
				expected[expectedIdx] = elt
				expectedIdx++
			}

		}
	}

	return expected, nil
}

func (o PositionAfter) GetExpected(entries []Movable, existing []Movable) ([]Movable, error) {
	return getPivotMovement(entries, existing, o.Pivot, o.Directly, ActionWhereAfter)
}

func (o PositionAfter) Move(entries []Movable, existing []Movable) ([]MoveAction, error) {
	expected, err := o.GetExpected(entries, existing)
	if err != nil {
		if errors.Is(err, errNoMovements) {
			return nil, nil
		}
		return nil, err
	}

	actions, err := GenerateMovements(existing, expected, entries, ActionWhereAfter, o.Pivot, o.Directly)
	if err != nil {
		return nil, err
	}

	return OptimizeMovements(existing, expected, entries, actions, o), nil
}

func (o PositionBefore) GetExpected(entries []Movable, existing []Movable) ([]Movable, error) {
	return getPivotMovement(entries, existing, o.Pivot, o.Directly, ActionWhereBefore)
}

func (o PositionBefore) Move(entries []Movable, existing []Movable) ([]MoveAction, error) {
	expected, err := o.GetExpected(entries, existing)
	if err != nil {
		if errors.Is(err, errNoMovements) {
			return nil, nil
		}
		return nil, err
	}

	slog.Debug("PositionBefore.Move()", "existing", existing, "expected", expected, "entries", entries)

	actions, err := GenerateMovements(existing, expected, entries, ActionWhereBefore, o.Pivot, o.Directly)
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
	for entry, idx := range *idxMap {
		if entry == moved {
			continue
		}

		if startingIdx > targetIdx && idx >= targetIdx {
			(*idxMap)[entry] = idx + 1
		} else if startingIdx < targetIdx && idx >= startingIdx && idx <= targetIdx {
			(*idxMap)[entry] = idx - 1
		}
	}
}

func OptimizeMovements(existing []Movable, expected []Movable, entries []Movable, actions []MoveAction, position Position) []MoveAction {
	simulated := make([]Movable, len(existing))
	copy(simulated, existing)

	simulatedIdxMap := createIdxMapFor(simulated)
	expectedIdxMap := createIdxMapFor(expected)

	var optimized []MoveAction

	switch position.(type) {
	case PositionBefore, PositionAfter:
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
			targetIdx = simulatedIdxMap[action.Destination] - 1
		case ActionWhereAfter:
			targetIdx = simulatedIdxMap[action.Destination] + 1
		}

		if targetIdx != currentIdx {
			optimized = append(optimized, action)
			simulatedIdxMap[action.Movable] = targetIdx
			updateSimulatedIdxMap(&simulatedIdxMap, action.Movable, currentIdx, targetIdx)
		}
	}

	slog.Debug("OptimizeMovements()", "optimized", optimized)

	return optimized
}

func GenerateMovements(existing []Movable, expected []Movable, entries []Movable, movement ActionWhereType, pivot Movable, directly bool) ([]MoveAction, error) {
	if len(existing) != len(expected) {
		return nil, ErrSlicesNotEqualLength
	}

	entriesIdxMap := createIdxMapFor(entries)
	existingIdxMap := createIdxMapFor(existing)
	expectedIdxMap := createIdxMapFor(expected)

	var movements []MoveAction
	var previous Movable
	for _, elt := range entries {
		slog.Debug("GenerateMovements()", "elt", elt, "existing", existingIdxMap[elt], "expected", expectedIdxMap[elt], "len(expected)", len(expected))
		// If existing index for the element matches the expected one, skip it over
		if existingIdxMap[elt] == expectedIdxMap[elt] {
			continue
		}

		if previous != nil {
			movements = append(movements, MoveAction{
				Movable:     elt,
				Destination: previous,
				Where:       ActionWhereAfter,
			})
			previous = elt
			continue
		}
		if expectedIdxMap[elt] == 0 {
			movements = append(movements, MoveAction{
				Movable:     elt,
				Destination: nil,
				Where:       ActionWhereTop,
			})
			previous = elt
		} else if expectedIdxMap[elt] == len(expectedIdxMap)-1 {
			movements = append(movements, MoveAction{
				Movable:     elt,
				Destination: nil,
				Where:       ActionWhereBottom,
			})
			previous = elt
		} else {
			var where ActionWhereType

			var pivot Movable
			switch movement {
			case ActionWhereBottom:
				where = ActionWhereBottom
			case ActionWhereAfter:
				pivot = expected[expectedIdxMap[elt]-1]
				where = ActionWhereAfter
			case ActionWhereTop:
				pivot = existing[0]
				where = ActionWhereBefore
			case ActionWhereBefore:
				eltExpectedIdx := expectedIdxMap[elt]
				pivot = expected[eltExpectedIdx+1]
				where = ActionWhereBefore
				// if previous was nil (we are processing the first element in entries set)
				// and selected pivot is part of the entries set it means the order of entries
				// changes between existing and expected sets. If direct move has been requested,
				// we need to find the correct pivot point for the move.
				if _, ok := entriesIdxMap[pivot]; ok && directly {
					// The actual pivot for the move is the element that follows all elements
					// from the existing set.
					pivotIdx := eltExpectedIdx + len(entries)
					if pivotIdx >= len(expected) {
						// This should never happen as by definition there is at least
						// element (pivot point) at the end of the expected slice.
						return nil, ErrInvalidMovementPlan
					}
					pivot = expected[pivotIdx]
				}
			}

			movements = append(movements, MoveAction{
				Movable:     elt,
				Destination: pivot,
				Where:       where,
			})
			previous = elt
		}

	}

	slog.Debug("GeneraveMovements()", "movements", movements)

	return movements, nil
}

func (o PositionTop) GetExpected(entries []Movable, existing []Movable) ([]Movable, error) {
	entriesIdxMap := createIdxMapFor(entries)

	filtered := removeEntriesFromExisting(existing, func(entry Movable) bool {
		_, ok := entriesIdxMap[entry]
		return ok
	})

	expected := append(entries, filtered...)

	return expected, nil
}

func (o PositionTop) Move(entries []Movable, existing []Movable) ([]MoveAction, error) {
	expected, err := o.GetExpected(entries, existing)
	if err != nil {
		return nil, err
	}
	actions, err := GenerateMovements(existing, expected, entries, ActionWhereTop, nil, false)
	if err != nil {
		return nil, err
	}

	return OptimizeMovements(existing, expected, entries, actions, o), nil
}

func (o PositionBottom) GetExpected(entries []Movable, existing []Movable) ([]Movable, error) {
	entriesIdxMap := createIdxMapFor(entries)

	filtered := removeEntriesFromExisting(existing, func(entry Movable) bool {
		_, ok := entriesIdxMap[entry]
		return ok
	})

	expected := append(filtered, entries...)

	return expected, nil
}

func (o PositionBottom) Move(entries []Movable, existing []Movable) ([]MoveAction, error) {
	slog.Debug("PositionBottom.Move())", "entries", entries, "existing", existing)
	expected, err := o.GetExpected(entries, existing)
	if err != nil {
		return nil, err
	}

	actions, err := GenerateMovements(existing, expected, entries, ActionWhereBottom, nil, false)
	if err != nil {
		return nil, err
	}
	return OptimizeMovements(existing, expected, entries, actions, o), nil
}

type Movement struct {
	Entries  []Movable
	Position Position
}

func MoveGroups(existing []Movable, movements []Movement) ([]MoveAction, error) {
	expected := existing
	for idx := range len(movements) - 1 {
		position := movements[idx].Position
		entries := movements[idx].Entries
		slog.Debug("MoveGroups()", "position", position, "existing", existing, "entries", entries)
		result, err := position.GetExpected(entries, expected)
		if err != nil {
			if !errors.Is(err, errNoMovements) {
				return nil, err
			}
			continue
		}
		expected = result
	}

	entries := movements[len(movements)-1].Entries
	position := movements[len(movements)-1].Position
	slog.Debug("MoveGroups()", "position", position, "expected", expected, "entries", entries)
	return position.Move(entries, expected)
}

func MoveGroup(position Position, entries []Movable, existing []Movable) ([]MoveAction, error) {
	return position.Move(entries, existing)
}

type Move struct {
	Position Position
	Existing []Movable
}
