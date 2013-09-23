package main

var usage = `
dep is a tool for the installation and update of Go packages

It prevents breakage of existing packages in GOPATH with the help
of a tentative installation in a temporary GOPATH and by detecting
breakage of dependancies in the go dependancy format (GDF).

Packages that use relative import paths are not supported and might
break.

For more information, see http://github.com/metakeule/dep

Required environment variables:
       - GOPATH should point to a valid GOPATH
       - DEP_TMP should point to a directory where tentative
         and temporary installations go to

PLEASE BE WARNED:
All actions act within the current GOPATH environment.
As dep is experimental at this point, you might loose all
your packages. No guarantee is made for whatever.

Usage:

         dep [options] command [package] 

If no package is given the package of the current working directory
is chosen.

Options:
    -verbose          Print details about the actions taken.
    -json             Print in machine readable json format

The commands are:

    gdf          Print the package's GDF.

    get          go get -u the given package and its dependancies
                 without breaking installed packages. Returns a list
                 of incompatibilities if there were any.
                 You should check, the integrity of your GOPATH with 'dep check'
                 before running 'dep get', otherwise dependencies might not be
                 checked properly.               
    
    track        track the imported packages with their revisions in 
                 the dep-rev.json file inside the package directory
                 That file will be used to get the exact same revisions
                 when using dep get.
    
    register     Add / update package's GDF inside the registry. 
                 Only needed for packages in the GOPATH that had already
                 been installed with other tools (e.g. go get / go install).
                 Not needed for packages that were installed via dep get.
    
    register-all like register, but also registers any included packages, as
                 they were currently in GOPATH/src
    
    unregister   removes a package from the registry
    
    diff         Show the difference in the GDFs between the given package 
                 and its GDF as it is in the registry.
    
    lint         Check if the given package respects the recommendations
                 for a package maintainer as given by the GDF.
                 Please keep in mind that not all recommendations can be
                 automatically checked.
    
    init         (Re)initialize the registry for the whole GOPATH and
                 check for incompatibilities in exports between the packages 
                 in GOPATH/src. WARNING: this erases the former compatibility
                 informations in the registry and the checksums of the working
                 init functions.
    
    check        checks the integrity of the whole GOPATH while respecting the
                 current registry.
`