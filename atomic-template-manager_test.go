package atm

import (
	"os"
	"testing"
)

func TestMeetsInterface(t *testing.T) {
	var _ Manager = New()
}

func TestErrorOnExecutingTemplate(t *testing.T) {
	var m Manager = New()

	err := m.ExecuteTemplate(os.Stdin, "testing", nil)

	if err == nil {
		t.Fatal("Expected error but got none.")
	}

	err = m.ExecuteTemplate(os.Stdin, "root", nil)

	if err == nil {
		t.Fatal("Expected error but got none.")
	}
}
