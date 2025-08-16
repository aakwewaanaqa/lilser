package urling

type QueryMode int

const (
	// Normal Do nothing to the query value
	Normal QueryMode = 0
	// TrimSpace Trim the space in the query value
	TrimSpace QueryMode = 1
	// NoEmpty Do not add the query value if it is empty
	NoEmpty QueryMode = 2
	// ErrorEmpty Throw error if the query value is empty
	ErrorEmpty QueryMode = 4
)
