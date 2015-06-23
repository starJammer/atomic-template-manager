//Package atm contains a template manager that tries to manage html.Template.
/*
Here is some sample code from the README.md file. Assume this is the directory
structure.

    /tmp/template-dir/
    -----------------/template-1.html
    -----------------/template-2.html
    -----------------/atoms/
    -----------------------/atom-1.html
    -----------------------/atom-2.tpl  //this one is a tpl file extension
    -----------------------/subatoms/sub-atom-1.html
    -----------------/pages/
    -----------------------/page-1.html

Now you can do:

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
*/
package atm
