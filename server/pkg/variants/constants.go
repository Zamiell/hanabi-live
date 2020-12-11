package variants

// iota starts at 0 and counts upwards
// i.e. StackDirectionUndecided = 0, StackDirectionUp = 1, etc.

const (
	StackDirectionUndecided = iota
	StackDirectionUp
	StackDirectionDown
	StackDirectionFinished
)

const (
	DefaultVariantName = "No Variant"

	// The "Up or Down" variants have "START" cards.
	// Rank 0 is the stack base.
	// Rank 1-5 are the normal cards.
	// Rank 6 is a card of unknown rank.
	// Rank 7 is a "START" card.
	StartCardRank = 7
)

const (
	pointsPerStack = 5

	// A "reversed" version of every suit exists.
	suitReversedSuffix = " Reversed"
)
