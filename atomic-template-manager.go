package atm

import (
	"errors"
	ht "html/template"
	"io"
	"io/ioutil"
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
	//AddDirectories will add a base directory to be scanned for templates
	//Any future directories you add SHOULD NOT be a descendant of a directory
	//that was previously added. Call ParseTemplates to parse templates in the
	//directories
	AddDirectories(dirs ...string) (Manager, error)
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
	//The defaults are {{ and }}. Call this before calling ParseTemplates
	Delims(left, right string) Manager
	//Funcs sets the FuncMap for all the templates
	Funcs(funcMap ht.FuncMap) Manager
	//Lookup finds a template by name
	Lookup(name string) *ht.Template
	//ParseTemplates parses all templates found in the directories
	//added by AddDirectories calls and any directories passed in here
	//Any errors encountered during reading the files are returned
	//in the slice of errors
	//
	//If you wish to update the template definitions, because
	//you are writing new templates during http requests,
	//call ParseTemplates again with no arguments. It will reparse
	//all the template directories
	ParseTemplates() []error

	//SetReparseOnExecute will tell the manager whether or not
	//you want to cache the templates. set to true if you're
	//developing and want to see changes in your templates immediately
	//false otherwise because it will take longer
	//
	//Because of the implementation details of the html.Template
	//it is best to either set this to true or set this to false
	//and leave it alone. Flipping it back and forth is not great.
	//also, because templates might be tied to each other,
	//setting this to false will cause the ExecuteTemplate method
	//to be almost like calling ParseTemplates every time
	//because the entire template hierarchy has to be recreated
	//each time.
	SetReparseOnExecute(reparse bool) Manager

	//Templates returns the number of templates in the manager
	Templates() []*ht.Template
}

type manager struct {
	rootex     *sync.Mutex
	root       *ht.Template
	funcMap    ht.FuncMap
	dirs       map[string]bool
	extensions map[string]bool
	templates  []*ht.Template
	reparse    bool

	leftDelim, rightDelim string
}

func (m *manager) AddDirectories(dirs ...string) (Manager, error) {
	//add incoming directories to list
	for _, v := range dirs {
		abs, err := filepath.Abs(v)
		if err != nil {
			return m, err
		}
		m.dirs[abs] = true
	}
	return m, nil
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
	if m.reparse {
		m.ParseTemplates()
	}

	m.rootex.Lock()
	defer m.rootex.Unlock()

	return m.root.ExecuteTemplate(wr, name, data)
}

func (m *manager) Delims(left, right string) Manager {
	m.leftDelim, m.rightDelim = left, right
	return m
}

func (m *manager) Funcs(funcMap ht.FuncMap) Manager {
	m.funcMap = funcMap
	return m
}

func (m *manager) Lookup(name string) *ht.Template {
	m.rootex.Lock()
	defer m.rootex.Unlock()
	return m.root.Lookup(name)
}

func (m *manager) ParseTemplates() []error {

	m.rootex.Lock()
	if len(m.root.Templates()) > 0 {
		m.root = ht.New("atomic-template-manager")
		m.templates = make([]*ht.Template, 0)
	}
	m.root.Delims(m.leftDelim, m.rightDelim)
	m.root.Funcs(m.funcMap)
	m.rootex.Unlock()

	var c = make(chan error)
	var w sync.WaitGroup

	//the function we'll use for walking each directory
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
			ext = strings.TrimPrefix(filepath.Ext(info.Name()), ".")

			//if the file extension matches any file extension
			//we're looking for then parse it and add it
			if _, ok := m.extensions[ext]; ok {
				alias := templateAliases(dir, path, ext)
				lalias := len(alias)
				if lalias == 0 {
					return nil
				}
				var newTemplate *ht.Template
				//create the new template using the first alias
				m.rootex.Lock()
				newTemplate = m.root.New(alias[0])
				fileContents, err := ioutil.ReadFile(path)

				if err != nil {
					return err
				}

				_, err = newTemplate.Parse(string(fileContents))

				if err != nil {
					return err
				}

				m.templates = append(m.templates, newTemplate)

				for i := 1; i < len(alias); i++ {
					_, err = m.root.AddParseTree(alias[i], newTemplate.Tree)
					if err != nil {
						return err
					}
				}
				m.rootex.Unlock()

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

func (m *manager) SetReparseOnExecute(reparse bool) Manager {
	m.reparse = reparse
	return m
}

func (m *manager) Templates() []*ht.Template {
	return m.templates
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
	aliasWithExtension := strings.TrimPrefix(path, root+"/")
	aliasWithoutExtension := strings.TrimSuffix(aliasWithExtension, "."+ext)
	parts := strings.Split(aliasWithoutExtension, string(os.PathSeparator))

	if len(parts) < 1 {
		panic("Root and path are the same ( root = " + root + ", path = " + path + " )")
	}

	alias = append(alias, aliasWithExtension)

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
	man.root = ht.New("atomic-template-manager")
	man.dirs = make(map[string]bool)
	man.extensions = make(map[string]bool)
	man.extensions["html"] = true
	man.extensions["tpl"] = true
	man.reparse = false
	man.templates = make([]*ht.Template, 0)
	man.rootex = new(sync.Mutex)
	return man
}
