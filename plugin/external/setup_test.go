package external

import (
	"testing"

	"github.com/mholt/caddy"
)

func TestSetup(t *testing.T) {
	tests := []struct {
		input        string
		shouldErr    bool
		expectedZone string
	}{
		{`external`, false, ""},
		{`external example.org`, false, "example.org."},
	}

	for i, test := range tests {
		c := caddy.NewTestController("dns", test.input)
		e, err := parse(c)

		if test.shouldErr && err == nil {
			t.Errorf("Test %d: Expected error but found %s for input %s", i, err, test.input)
		}

		if err != nil {
			if !test.shouldErr {
				t.Errorf("Test %d: Expected no error but found one for input %s. Error was: %v", i, test.input, err)
			}
		}

		if !test.shouldErr && test.expectedZone != "" {
			if test.expectedZone != e.Zones[0] {
				t.Errorf("Test %d, expected zone %q for input %s, got: %q", i, test.expectedZone, test.input, e.Zones[0])
			}
		}
	}
}