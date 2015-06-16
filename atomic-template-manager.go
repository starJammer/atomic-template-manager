package atm

import (
	"errors"
	ht "html/template"
	"io"
)

var (
	NotADirErr = errors.New("This is not a directory.")
)

type Manager interface {
	//AddDirectory will add a base directory to be scanned for templates
	//Any future directories you add SHOULD NOT be a descendant of a directory
	//that was previously added. Call ParseDirs to parse templates in the
	//directories
	AddDirectory(dir string)
	//ExecuteTemplate will execute a template.
	ExecuteTemplate(wr io.Writer, name string, data interface{}) error
	//Delims Sets the delimiters to be used when parsing templates.
	//The defaults are {{ and }}. Call this before calling ParseDirs
	Delims(left, right string) Manager
	//Funcs sets the FuncMap for all the templates
	Funcs(funcMap ht.FuncMap) Manager
	//Lookup finds a template by name
	Lookup(name string) *ht.Template
	//ParseDirs parses all templates found in the directories
	//added by AddDirectory calls and any directories passed in here
	ParseDirs(dirs ...string) error
}

type manager struct {
	root *ht.Template
	dirs []string
}

func (m *manager) AddDirectory(dir string) {
	m.dirs = append(m.dirs, dir)
}

func (m *manager) ExecuteTemplate(wr io.Writer, name string, data interface{}) error {

	template := m.Lookup(name)
	return template.Execute(wr, data)
}

func (m *manager) Delims(left, right string) Manager {
	m.root.Delims(left, right)
	return m
}

func (m *manager) Funcs(funcMap ht.FuncMap) Manager {
	m.root.Funcs(funcMap)
	return m
}

func (m *manager) Lookup(name string) *ht.Template {
	return m.root.Lookup(name)
}

func (m *manager) ParseDirs(dirs ...string) error {

	return nil
}

func New() Manager {
	man := new(manager)
	man.root = ht.New("root")
	return man
}
