package movement_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"movements/movement"
)

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

				Expect(moves[0].EntryName).To(Equal("A"))
				Expect(moves[0].Where).To(Equal("top"))
				Expect(moves[0].Destination).To(Equal("top"))

				Expect(moves[1].EntryName).To(Equal("B"))
				Expect(moves[1].Where).To(Equal("after"))
				Expect(moves[1].Destination).To(Equal("A"))

				Expect(moves[2].EntryName).To(Equal("C"))
				Expect(moves[2].Where).To(Equal("after"))
				Expect(moves[2].Destination).To(Equal("B"))
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

				Expect(moves[0].EntryName).To(Equal("E"))
				Expect(moves[0].Where).To(Equal("after"))
				Expect(moves[0].Destination).To(Equal("D"))
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
				It("should generate only move commands to sort entries", func() {
					// A B C D E -> A C B D E
					entries := asMovable([]string{"C", "B"})
					moves, err := movement.MoveGroup(
						movement.PositionBefore{Directly: false, Pivot: Mock{"D"}},
						entries, existing,
					)

					Expect(err).ToNot(HaveOccurred())
					Expect(moves).To(HaveLen(2))
					Expect(moves[0].EntryName).To(Equal("C"))
					Expect(moves[0].Where).To(Equal("after"))
					Expect(moves[0].Destination).To(Equal("A"))

					Expect(moves[1].EntryName).To(Equal("B"))
					Expect(moves[1].Where).To(Equal("after"))
					Expect(moves[1].Destination).To(Equal("C"))
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

				Expect(moves[0].EntryName).To(Equal("A"))
				Expect(moves[0].Where).To(Equal("after"))
				Expect(moves[0].Destination).To(Equal("C"))

				Expect(moves[1].EntryName).To(Equal("B"))
				Expect(moves[1].Where).To(Equal("after"))
				Expect(moves[1].Destination).To(Equal("A"))
			})
		})
	})
})
