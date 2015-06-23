# Atomic Template Manager

## Installation

`go get github.com/starJammer/atomic-template-manager`

## Reasoning

Setting up a hierarchy of `html.Template`s in go can be annoying because you have to 
parse them on your own. I wanted something that I could setup with a directory and then
say Parse and have all my templates accessible in a nice naming convention. I chose
to use the naming convention from [Pattern Lab](http://patternlab.io/docs/pattern-including.html).
I'll show in the code below what I mean by the naming convention.

By the way, yes, I know that I could use `Template.ParseGlob( "" )` but this wasn't really what
I wanted.

## Usage

Suppose we have the following directory structure:

    /tmp/template-dir/
    -----------------/template-1.html
    -----------------/template-2.html
    -----------------/atoms/
    -----------------------/atom-1.html
    -----------------------/atom-2.tpl  //this one is a tpl file extension
    -----------------------/subatoms/sub-atom-1.html
    -----------------/pages/
    -----------------------/page-1.html

Now you can do the following:

```go
package main
import(
    //url is long but package name is actually atm
    "github.com/starJammer/atomic-template-manager" 
    "os"
    "fmt"
)

var man Manager

func main() {
    man = atm.New()

    //takes an unlimited number of strings that
    //should point to directories.
    man.AddDirectories( "/tmp/template-dir" )

    //Parse template produces a slice of errors. 
    //It tries to do as much as it can but you should
    //check to see if this slice is none nil
    errs := man.ParseTemplates()

	if errs != nil {
		fmt.Println( errs )
	}

    //We can execute one of the templates using a shorthand name
    //like in pattern lab.
    //Notice that there is no file extension
    //Additionally, the path separator becomes a hyphen
    man.ExecuteTemplate( os.Stdout, "pages-page-1", nil )

    //we can also use the full path to call the same template
    man.ExecuteTemplate( os.Stdout, "pages/page-1.html", nil )

    //Notice in the call below that the subatoms subdirectory is omitted from
    //the template name
    man.ExecuteTemplate( os.Stdout, "atoms-sub-atom-1", nil )

    //we can also just use the long name
    man.ExecuteTemplate( os.Stdout, "atoms/subatoms/sub-atom-1.html", nil )

    //The manager scans for both tpl and html files so we can
    //still call atoms-2 with:
    man.ExecuteTemplate( os.Stdout, "atoms-atom-2", nil )

    //anything at the top directory is called with either shorthand
    man.ExecuteTemplate( os.Stdout, "template-1", nil )

    //OR the full filename
    man.ExecuteTemplate( os.Stdout, "template-1.html", nil )

    //use a custom file extension. Call this before ParseTemplates()
    man.AddFileExtension( "new-tpl" )
    //Ignore all files with html extension. Call this before ParseTemplates()
    man.RemoveFileExtension( "html" )
    //FYI, using files with no extension is not supported.

    //same as html/template package
    man.Delims( "<<<", ">>>" )

    //same to html.Template.Funcs
    man.Funcs( map[string]interface{}{
        "func1" : func( arg string ) string {return arg}
        "func2" : func( arg string ) string {return arg+"2"}
    })

    //Same as html/template package
    man.Lookup( "template-name" )

    //Same as html/template package
    man.Templates()

    //reparse all templates on ExecuteTemplate, ok for dev not for prod
    man.SetReparseOnExecute(true)
}
```

## Disadvantages

### 1 
The way `html.Template` works means that the whole template hierarchy needs to be 
reparsed if you want to "watch" for changes by setting `SetReparseOnExecute( true )`.
`html.Template` doesn't easily and directly allow one template definition
definition to be swapped out for another under the same name. For example, it doesn't
allow 

    var template *html.Template
    template.New( "some-template" ).Parse( "hello world" )
    template.Replace( "some-template" ).Parse( "hello universe" )

Oh well.

### 2
Also, even though you can add multiple directories to be parsed, all the templates 
from all directories are added to the same hierarchy so you can't have any duplicate
templates across directories. You can't have:

    /tmp/rootdir1/atom/test-1.html
    /tmp/rootdir2/atom/test-1.html

You'll get some error....well, I haven't tested this but I know it'll break.

### 3
I haven't implemented the part where you can add %%- in front of files and directories,
as in `00-directory/01-template-1.html`. If I get to this I'll get to it. The numbers
were used in pattern lab for specifying order of templates in a gui but I didn't really see
too much use for this feature here. Let me know if you want it. The function is already
there, I just have to write it.

## Miscellaneous

I tried to make it safe from concurrent goroutines by using a mutex. I also tried
to make the parsing of the directories semi-concurrent. That is to say, different root
directories will be parse concurrently but each root directory will be parsed
sequentially.
