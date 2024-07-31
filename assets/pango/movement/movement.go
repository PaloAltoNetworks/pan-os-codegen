package movement

import (
	"fmt"
	"log/slog"
	"slices"
)

var _ = slog.LevelDebug

type Movable interface {
	EntryName() string
}

type MoveAction struct {
	EntryName   string
	Where       string
	Destination string
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

func processPivotMovement(entries []Movable, existing []Movable, pivot Movable, direct bool, movement movementType) ([]MoveAction, error) {
	existingLen := len(existing)
	existingIdxMap := make(map[Movable]int, existingLen)

	for idx, elt := range existing {
		existingIdxMap[elt] = idx
	}

	pivotIdx := findPivotIdx(existing, pivot)
	if pivotIdx == -1 {
		return nil, fmt.Errorf("pivot point not found in the list of existing items")
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
				if existingEntryIdx >= pivotIdx {
					movementRequired = true
					break
				}
			case movementAfter:
				if existingEntryIdx <= pivotIdx {
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
			return nil, nil
		}
	}

	expected := make([]Movable, len(existing))

	entriesIdxMap := make(map[Movable]int, len(entries))
	for idx, elt := range entries {
		entriesIdxMap[elt] = idx
	}

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

	return GenerateMovements(existing, expected, entries)
}

func (o PositionAfter) Move(entries []Movable, existing []Movable) ([]MoveAction, error) {
	return processPivotMovement(entries, existing, o.Pivot, o.Directly, movementAfter)
}

func (o PositionBefore) Move(entries []Movable, existing []Movable) ([]MoveAction, error) {
	return processPivotMovement(entries, existing, o.Pivot, o.Directly, movementBefore)
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

func longestCommonSubsequence(S []Movable, T []Movable) [][]Movable {

	r := len(S)
	n := len(T)

	L := make([][]int, r)
	for idx := range len(T) {
		L[idx] = make([]int, n)
	}
	z := 0

	var results [][]Movable

	for i := 0; i < r; i++ {
		for j := 0; j < n; j++ {
			if S[i].EntryName() == T[j].EntryName() {
				if i == 0 || j == 0 {
					L[i][j] = 1
				} else {
					L[i][j] = L[i-1][j-1] + 1
				}

				if L[i][j] > z {
					slog.Debug("L[i][j] > z", "L[i][j]", L[i][j], "z", z, "i-z", i-z, "i", i)
					results = nil
					results = append(results, S[i-z:i+1])
					z = L[i][j]
					slog.Debug("L[i][j] > z", "results", results)
				} else if L[i][j] == z {
					results = append(results, S[i-z:i+1])
					slog.Debug("L[i][j] == z", "i-z", i, "i", i+1)
				}
				slog.Debug("Still", "results", results)
			} else {
				L[i][j] = 0
			}
		}
	}

	slog.Debug("commonSubsequence", "results", results)

	return results
}

func GenerateMovements(existing []Movable, expected []Movable, entries []Movable) ([]MoveAction, error) {
	if len(existing) != len(expected) {
		return nil, fmt.Errorf("existing length != expected length: %d != %d", len(existing), len(expected))
	}

	common := longestCommonSubsequence(existing, expected)

	entriesIdxMap := make(map[Movable]int, len(entries))
	for idx, elt := range entries {
		entriesIdxMap[elt] = idx
	}

	var commonSequence []Movable
	for _, elt := range common {
		filtered := removeEntriesFromExisting(elt, func(elt Movable) bool {
			_, ok := entriesIdxMap[elt]
			return ok
		})

		if len(filtered) > len(commonSequence) {
			commonSequence = filtered
		}

	}

	existingIdxMap := make(map[Movable]int, len(existing))
	for idx, elt := range existing {
		existingIdxMap[elt] = idx
	}

	expectedIdxMap := make(map[Movable]int, len(expected))
	for idx, elt := range expected {
		expectedIdxMap[elt] = idx
	}

	commonLen := len(commonSequence)
	commonIdxMap := make(map[Movable]int, len(commonSequence))
	for idx, elt := range commonSequence {
		commonIdxMap[elt] = idx
	}

	var movements []MoveAction

	var previous Movable
	for _, elt := range entries {
		slog.Debug("GenerateMovements", "elt", elt.EntryName(), "existingIdx", existingIdxMap[elt], "expectedIdx", expectedIdxMap[elt])
		if existingIdxMap[elt] == expectedIdxMap[elt] {
			continue
		}

		if expectedIdxMap[elt] == 0 {
			movements = append(movements, MoveAction{
				EntryName:   elt.EntryName(),
				Destination: "top",
				Where:       "top",
			})
			previous = elt
		} else if len(commonSequence) > 0 {
			if expectedIdxMap[elt] < expectedIdxMap[commonSequence[0]] {
				if previous == nil {
					previous = expected[0]
				}
				movements = append(movements, MoveAction{
					EntryName:   elt.EntryName(),
					Destination: previous.EntryName(),
					Where:       "after",
				})
				previous = elt
			} else if expectedIdxMap[elt] > expectedIdxMap[commonSequence[commonLen-1]] {
				if previous == nil {
					previous = commonSequence[commonLen-1]
				}
				movements = append(movements, MoveAction{
					EntryName:   elt.EntryName(),
					Destination: previous.EntryName(),
					Where:       "after",
				})
				previous = elt

			} else if expectedIdxMap[elt] > expectedIdxMap[commonSequence[0]] {
				if previous == nil {
					previous = commonSequence[0]
				}
				movements = append(movements, MoveAction{
					EntryName:   elt.EntryName(),
					Destination: previous.EntryName(),
					Where:       "after",
				})
				previous = elt
			}
		} else {
			movements = append(movements, MoveAction{
				EntryName:   elt.EntryName(),
				Destination: previous.EntryName(),
				Where:       "after",
			})
			previous = elt
		}

		slog.Debug("GenerateMovements()", "existing", existingIdxMap[elt], "expected", expectedIdxMap[elt])
	}

	_ = previous

	slog.Debug("GenerateMovements()", "movements", movements)

	return movements, nil
}

func (o PositionTop) Move(entries []Movable, existing []Movable) ([]MoveAction, error) {
	entriesIdxMap := make(map[Movable]int, len(entries))
	for idx, elt := range entries {
		entriesIdxMap[elt] = idx
	}

	filtered := removeEntriesFromExisting(existing, func(entry Movable) bool {
		_, ok := entriesIdxMap[entry]
		return ok
	})

	expected := append(entries, filtered...)

	return GenerateMovements(existing, expected, entries)
}

func (o PositionBottom) Move(entries []Movable, existing []Movable) ([]MoveAction, error) {
	entriesIdxMap := make(map[Movable]int, len(entries))
	for idx, elt := range entries {
		entriesIdxMap[elt] = idx
	}

	filtered := removeEntriesFromExisting(existing, func(entry Movable) bool {
		_, ok := entriesIdxMap[entry]
		return ok
	})

	expected := append(filtered, entries...)

	return GenerateMovements(existing, expected, entries)
}

func MoveGroup(position Position, entries []Movable, existing []Movable) ([]MoveAction, error) {
	return position.Move(entries, existing)
}
