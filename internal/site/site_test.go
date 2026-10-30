package site

import "testing"

func TestParseAcceptsConfiguredSiteNames(t *testing.T) {
	for _, test := range []struct {
		input string
		want  ID
	}{
		{input: "FluxA", want: FluxA},
		{input: "fluxa", want: FluxA},
		{input: "Charity", want: Charity},
		{input: "charity", want: Charity},
	} {
		got, ok := Parse(test.input)
		if !ok || got != test.want {
			t.Fatalf("Parse(%q) = %q, %v; want %q, true", test.input, got, ok, test.want)
		}
	}
}

func TestParseRejectsUnknownSite(t *testing.T) {
	if got, ok := Parse("unknown"); ok || got != "" {
		t.Fatalf("Parse(unknown) = %q, %v; want empty, false", got, ok)
	}
}
