package site

import "strings"

type ID string

const (
	FluxA   ID = "FluxA"
	Charity ID = "Charity"
)

func Parse(value string) (ID, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "fluxa":
		return FluxA, true
	case "charity":
		return Charity, true
	default:
		return "", false
	}
}

func All() []ID {
	return []ID{FluxA, Charity}
}
