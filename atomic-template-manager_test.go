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

func createTestTemplates(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	err = writeTemplateFile(dir+"/templates/top-level.html", ``)
	if err != nil {
		t.Fatal(err)
	}

	err = writeTemplateFile(dir+"/templates/none.none", ``)
	if err != nil {
		t.Fatal(err)
	}

	err = writeTemplateFile(dir+"/templates/atoms/atom-1.html", ``)
	if err != nil {
		t.Fatal(err)
	}

	err = writeTemplateFile(dir+"/templates/atoms/atom-2.tpl", ``)
	if err != nil {
		t.Fatal(err)
	}

	err = writeTemplateFile(dir+"/templates/atoms/fonts/font-1.html", ``)
	if err != nil {
		t.Fatal(err)
	}
}

func writeTemplateFile(path, contents string) error {
	templateFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer templateFile.Close()
	_, err = templateFile.WriteString(contents)
	if err != nil {
		return err
	}
	return nil
}

func destroyAll(t *testing.T) {
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
	defer destroyAll(t)

	var man Manager = New()
	man.AddDirectories("./templates")
	man.ParseTemplates()
	if len(man.Templates()) > 0 {
		t.Fatal("There were templates even though the directories were empty.")
	}
}

func TestDefaultTemplatesAreFound(t *testing.T) {
	createDirs(t)
	createTestTemplates(t)
	defer destroyAll(t)

	var man Manager = New()

	man.AddDirectories("./templates")
	man.ParseTemplates()
	if len(man.Templates()) != 4 {
		t.Fatalf("We expected 4 templates but had : %d", len(man.Templates()))
	}
}
