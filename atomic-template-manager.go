package atm

import (
	"errors"
	ht "html/template"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	NotADirErr          = errors.New("This is not a directory.")
	TemplateNotFoundErr = errors.New("The template could not be found.")
)

type Manager interface {
	//AddDirectory will add a base directory to be scanned for templates
	//Any future directories you add SHOULD NOT be a descendant of a directory
	//that was previously added. Call ParseDirs to parse templates in the
	//directories
	AddDirectory(dir string) Manager
	//AddFileExtension adds a file extension that will be considered
	//a template. By default, both .html and .tpl will be considered
	//templates.
	AddFileExtension(ext string) Manager
	//RemoveFileExtension removes an ext so it isn't considered a
	//template. Use this to remove the default extensions
	RemoveFileExtension(ext string) Manager

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
	//Any errors encountered during reading the files are returned
	//in the slice of errors
	ParseDirs(dirs ...string) []error
}

type manager struct {
	root       *ht.Template
	dirs       map[string]bool
	extensions map[string]bool
	aliases    map[string]*string
	templates  map[string]*ht.Template
}

func (m *manager) AddDirectory(dir string) Manager {
	m.dirs[dir] = true
	return m
}

func (m *manager) AddFileExtension(ext string) Manager {
	m.extensions[ext] = true
	return m
}

func (m *manager) RemoveFileExtension(ext string) Manager {
	delete(m.extensions, ext)
	return m
}

func (m *manager) ExecuteTemplate(wr io.Writer, name string, data interface{}) error {
	return m.root.ExecuteTemplate(wr, name, data)
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

func (m *manager) ParseDirs(dirs ...string) []error {
	//add incoming directories to list
	for _, v := range dirs {
		m.dirs[v] = true
	}

	var c = make(chan error)
	var w sync.WaitGroup
	m.aliases = make(map[string]*string)
	m.templates = make(map[string]*ht.Template)

	var walkDir = func(dir string) {
		defer w.Done()
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err == os.ErrPermission {
				c <- errors.New(path + " " + err.Error())
				return nil
			}

			if info.IsDir() {
				return nil
			}
			var ext string

			ext = filepath.Ext(info.Name())

			//if the file extension matches any file extension
			//we're looking for then parse it and add it
			if _, ok := m.extensions[ext]; ok {
				alias := templateAliases(dir, path, ext)
				//use a string pointer to avoid having the same string floating around
				//just a small stupid attempt at optimization
				var pathPoint *string
				pathPoint = &path
				for _, v := range alias {
					m.aliases[v] = pathPoint
					m.templates[*pathPoint] = nil
				}
			}

			return nil
		})

		if err != nil {
			c <- err
		}
	}

	//start parsing the directories
	for d, _ := range m.dirs {
		w.Add(1)
		go walkDir(d)
	}

	go func() {
		w.Wait()
		close(c)
	}()

	var errors = make([]error, 0)
	for err := range c {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}

//templateAliases will generate the aliases that
//we will be able to use to include/access the
//template located by path.
//Root should be the root template directory
//so we can generate the aliases accordingly.
//
//Ex. Root = /tmp
//    Path = /tmp/atom/template-1.html
//    Aliases = { "atom-template-1", "atom/template-1" }
//Ex. Root = /tmp
//    Path = /tmp/atom/subdir/template-1.html
//    Aliases = { "atom-template-1", "atom/subdir/template-1" }
//Ex. Root = /tmp
//    Path = /tmp/00-atom/00-subdir/template-1.html
//    Aliases = { "atom-template-1", "00-atom/00-subdir/template-1" }
func templateAliases(root, path, ext string) []string {
	alias := make([]string, 0, 2)
	remainingPath := strings.TrimPrefix(path, root)
	remainingPath = strings.TrimSuffix(path, "."+ext)
	parts := strings.Split(string(os.PathSeparator), remainingPath)

	if len(parts) < 1 {
		panic("Root and path are the same ( root = " + root + ", path = " + path + " )")
	}

	alias = append(alias, remainingPath)

	if len(parts) == 1 {
		alias = append(alias, removeLeadingNumbers(parts[0]))
	} else {
		alias = append(alias, removeLeadingNumbers(parts[0])+"-"+removeLeadingNumbers(parts[len(parts)-1]))
	}
	return alias
}

func removeLeadingNumbers(p string) string {
	return p
}

func New() Manager {
	man := new(manager)
	man.root = ht.New("root")
	man.dirs = make(map[string]bool)
	man.extensions = make(map[string]bool)
	man.extensions["html"] = true
	man.extensions["tpl"] = true
	return man
}
