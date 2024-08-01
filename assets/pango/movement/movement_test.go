package movement_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"movements/movement"
)

var _ = fmt.Printf

type Mock struct {
	Name string
}

func (o Mock) EntryName() string {
	return o.Name
}

func asMovable(mocks []string) []movement.Movable {
	var movables []movement.Movable

	for _, elt := range mocks {
		movables = append(movables, Mock{elt})
	}

	return movables
}

var _ = Describe("LCS", func() {
	Context("with two common substrings", func() {
		existing := asMovable([]string{"A", "B", "C", "D", "E"})
		expected := asMovable([]string{"C", "A", "B", "D", "E"})
		It("should return two sequences of two elements", func() {
			options := movement.LongestCommonSubstring(existing, expected)
			Expect(options).To(HaveLen(2))

			Expect(options[0]).To(HaveExactElements(asMovable([]string{"A", "B"})))
			Expect(options[1]).To(HaveExactElements(asMovable([]string{"D", "E"})))
		})
	})
	// Context("with one very large common substring", func() {
	// 	It("should return one sequence of elements in a reasonable time", Label("benchmark"), func() {
	// 		var elts []string
	// 		elements := 50000
	// 		for idx := range elements {
	// 			elts = append(elts, fmt.Sprintf("%d", idx))
	// 		}
	// 		existing := asMovable(elts)
	// 		expected := existing

	// 		options := movement.LongestCommonSubstring(existing, expected)
	// 		Expect(options).To(HaveLen(1))
	// 		Expect(options[0]).To(HaveLen(elements))
	// 	})
	// })
})

var _ = Describe("Movement", func() {
	Context("With PositionTop used as position", func() {
		Context("when existing positions matches expected", func() {
			It("should generate no movements", func() {
				expected := asMovable([]string{"A", "B", "C"})
				moves, err := movement.MoveGroup(movement.PositionTop{}, expected, expected)
				Expect(err).ToNot(HaveOccurred())
				Expect(moves).To(HaveLen(0))
			})
		})
		Context("when it has to move two elements", func() {
			It("should generate three move actions", func() {
				entries := asMovable([]string{"A", "B", "C"})
				existing := asMovable([]string{"D", "E", "A", "B", "C"})

				moves, err := movement.MoveGroup(movement.PositionTop{}, entries, existing)
				Expect(err).ToNot(HaveOccurred())
				Expect(moves).To(HaveLen(3))

				Expect(moves[0].Movable.EntryName()).To(Equal("A"))
				Expect(moves[0].Where).To(Equal(movement.ActionWhereTop))
				Expect(moves[0].Destination).To(BeNil())

				Expect(moves[1].Movable.EntryName()).To(Equal("B"))
				Expect(moves[1].Where).To(Equal(movement.ActionWhereAfter))
				Expect(moves[1].Destination.EntryName()).To(Equal("A"))

				Expect(moves[2].Movable.EntryName()).To(Equal("C"))
				Expect(moves[2].Where).To(Equal(movement.ActionWhereAfter))
				Expect(moves[2].Destination.EntryName()).To(Equal("B"))
			})
		})
		Context("when expected order is reversed", func() {
			It("should generate required move actions to converge lists", func() {
				entries := asMovable([]string{"E", "D", "C", "B", "A"})
				existing := asMovable([]string{"A", "B", "C", "D", "E"})
				moves, err := movement.MoveGroup(movement.PositionTop{}, entries, existing)
				Expect(err).ToNot(HaveOccurred())

				Expect(moves).To(HaveLen(4))
			})
		})
	})
	Context("With PositionBottom used as position", func() {
		Context("when it needs to move one element", func() {
			It("should generate a single move action", func() {
				entries := asMovable([]string{"E"})
				existing := asMovable([]string{"A", "E", "B", "C", "D"})

				moves, err := movement.MoveGroup(movement.PositionBottom{}, entries, existing)
				Expect(err).ToNot(HaveOccurred())
				Expect(moves).To(HaveLen(1))

				Expect(moves[0].Movable.EntryName()).To(Equal("E"))
				Expect(moves[0].Where).To(Equal(movement.ActionWhereAfter))
				Expect(moves[0].Destination.EntryName()).To(Equal("D"))
			})
		})
	})

	Context("With PositionAfter used as position", func() {
		existing := asMovable([]string{"A", "B", "C", "D", "E"})
		Context("when direct position relative to the pivot is not required", func() {
			It("should not generate any move actions", func() {
				entries := asMovable([]string{"D", "E"})
				moves, err := movement.MoveGroup(
					movement.PositionAfter{Directly: false, Pivot: Mock{"B"}},
					entries, existing,
				)

				Expect(err).ToNot(HaveOccurred())
				Expect(moves).To(HaveLen(0))
			})
			Context("and moved entries are out of order", func() {
				FIt("should generate a single command to move B before D", func() {
					// A B C D E -> A B C E D
					entries := asMovable([]string{"E", "D"})
					moves, err := movement.MoveGroup(
						movement.PositionAfter{Directly: false, Pivot: Mock{"B"}},
						entries, existing,
					)

					Expect(err).ToNot(HaveOccurred())
					Expect(moves).To(HaveLen(1))

					Expect(moves[0].Movable.EntryName()).To(Equal("E"))
					Expect(moves[0].Where).To(Equal(movement.ActionWhereAfter))
					Expect(moves[0].Destination.EntryName()).To(Equal("C"))
				})
			})
		})

	})
	Context("With PositionBefore used as position", func() {
		existing := asMovable([]string{"A", "B", "C", "D", "E"})

		Context("when direct position relative to the pivot is not required", func() {
			Context("and moved entries are already before pivot point", func() {
				It("should not generate any move actions", func() {
					entries := asMovable([]string{"A", "B"})
					moves, err := movement.MoveGroup(
						movement.PositionBefore{Directly: false, Pivot: Mock{"D"}},
						entries, existing,
					)

					Expect(err).ToNot(HaveOccurred())
					Expect(moves).To(HaveLen(0))
				})
			})
			Context("and moved entries are out of order", func() {
				FIt("should generate a single command to move B before D", func() {
					// A B C D E -> A C B D E
					entries := asMovable([]string{"C", "B"})
					moves, err := movement.MoveGroup(
						movement.PositionBefore{Directly: false, Pivot: Mock{"D"}},
						entries, existing,
					)

					Expect(err).ToNot(HaveOccurred())
					Expect(moves).To(HaveLen(1))

					Expect(moves[0].Movable.EntryName()).To(Equal("B"))
					Expect(moves[0].Where).To(Equal(movement.ActionWhereAfter))
					Expect(moves[0].Destination.EntryName()).To(Equal("C"))
				})
			})
		})
		Context("when direct position relative to the pivot is required", func() {
			It("should generate required move actions", func() {
				// A B C D E -> C A B D E
				entries := asMovable([]string{"A", "B"})
				moves, err := movement.MoveGroup(
					movement.PositionBefore{Directly: true, Pivot: Mock{"D"}},
					entries, existing,
				)

				Expect(err).ToNot(HaveOccurred())
				Expect(moves).To(HaveLen(2))

				Expect(moves[0].Movable.EntryName()).To(Equal("A"))
				Expect(moves[0].Where).To(Equal(movement.ActionWhereBefore))
				Expect(moves[0].Destination.EntryName()).To(Equal("D"))

				Expect(moves[1].Movable.EntryName()).To(Equal("B"))
				Expect(moves[1].Where).To(Equal(movement.ActionWhereAfter))
				Expect(moves[1].Destination.EntryName()).To(Equal("A"))
			})
		})
	})
})
