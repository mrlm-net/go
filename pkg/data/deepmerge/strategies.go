package deepmerge

// SliceStrategy defines how slices should be merged.
type SliceStrategy uint8

const (
	SliceOverride SliceStrategy = iota // Replace (default)
	SliceAppend                        // Concatenate
	SliceUnion                         // Unique values
)

// String implements the Stringer interface.
func (s SliceStrategy) String() string {
	switch s {
	case SliceOverride:
		return "SliceOverride"
	case SliceAppend:
		return "SliceAppend"
	case SliceUnion:
		return "SliceUnion"
	default:
		return "SliceStrategy(unknown)"
	}
}

// MapStrategy defines how maps should be merged.
type MapStrategy uint8

const (
	MapOverride MapStrategy = iota // Replace (default)
	MapMerge                       // Recursive merge
)

// String implements the Stringer interface.
func (m MapStrategy) String() string {
	switch m {
	case MapOverride:
		return "MapOverride"
	case MapMerge:
		return "MapMerge"
	default:
		return "MapStrategy(unknown)"
	}
}
