package atm

import (
	"testing"
)

func TestMe(t *testing.T) {
	var _ Manager = New()
}
