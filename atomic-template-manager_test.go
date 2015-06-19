package atm

import (
	"os"
	"testing"
)

//create template directory and sub directories
func createDirs(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(dir+"/templates/atoms/fonts", os.ModeDir|os.ModePerm); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(dir+"/templates/pages/front-page", os.ModeDir|os.ModePerm); err != nil {
		t.Fatal(err)
	}
}
func destroyDirs(t *testing.T) {
	if err := os.RemoveAll("./templates"); err != nil {
		t.Fatal(err)
	}
}

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

func TestNoTemplatesAreFound(t *testing.T) {
	createDirs(t)
	defer destroyDirs(t)

	var man Manager = New()

	man.AddDirectory("./templates")
	if len(man.Templates()) > 0 {
		t.Fatal("There were templates even though the directories were empty.")
	}
}
