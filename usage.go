package main

var usage = `
dep is a tool for the installation and update of Go packages

It prevents breakage of existing packages in GOPATH with the help
of a tentative installation in a temporary GOPATH and by detecting
breakage of dependancies in the go dependancy format (GDF, see 
http://github.com/go-dep/gdf).

Packages that use relative import paths are not supported and might
break.

For more information, see http://github.com/go-dep/dep

Required environment variables:
  
  - GOPATH should point to a valid GOPATH
  - DEP_TMP should point to a directory where tentative
    and temporary installations go to

Package developers, please read: 
  https://github.com/go-dep/gdf/wiki/Recommendations-for-go-package-developers

PLEASE BE WARNED:
  
  All actions act within the current GOPATH environment.
  As dep is experimental at this point, you might loose all
  your packages. No guarantee is made for whatever.

Usage:

  dep [options] command [package] 

If no package is given the package of the current working directory
is chosen.

Options:
  
  -verbose            Print details about the actions taken.
  -json               Print in machine readable json format
  -y                  Answer all questions with 'yes'
  -no-warn            Suppress warnings
  -panic              Panic on errors
  -override           Pass a file with GDF definitions that should 
                      be taken instead of the registry (only for dep get)
  -skip-check         Skip the dep check at the beginning of the dep get. 
                      Use this only if you know what you are doing

The commands are:

  gdf                 Print the GDF of the package based on the local source dir

  get                 first does a 'dep check' for the GOPATH and then 
                      go get -u the given package and its dependancies
                      without breaking installed packages. 
                      Returns a list of incompatibilities if there were any.
                      If there are dep-rev.json files within the package or 
                      any of their imports, the revisions will be respected. 
                      That might lead to a "downgrade" of some packages. 
                      However at the end of a successful 'dep get' you will be 
                      shown a list of repositories that changed and from which 
                      revision to what. 
                      So you might decide on your own, if you want to keep the
                      changes.
                      If a file is passed with -override, it is considered a 
                      Json-Array of GDFs that replace the corresponding entries 
                      in the registry when doing error checking.
                      You can get a GDF of a package with 'dep gdf'.
                      Don't forget to register the changed packages 
                      after a successful update with dep get.
                      You will get a list of all differences between the GDF 
                      based on the source directory and the counterpart in the
                      registry by using 'dep check'.
                      'dep get' does not 'go install' the packages. So you might
                      want to do this after running 'dep get'.
                      Please be aware that even if no GDF compatibility has been 
                      broken, the updated/installed packages may be 
                      disfunctional.
                      So check if they work correctly (e.g. by running 
                      'go test ./...').
                      For every changed package repository you get a backup 
                      directory in the same folder. So you might simply rename 
                      them in case anything goes wrong.
                      Again, if everything is fine, don't forget to register the 
                      changes in the registry with 'dep register' or 
                      'dep register-included'.
  
  track               track the imported packages with their revisions in 
                      the dep-rev.json file inside the package directory
                      That file will be used to get the exact same revisions
                      when using dep get.
  
  register            Add / update package's GDF inside the registry.
                      
  imports             show the imported packages of the local package source dir 
  
  register-included   like register, but also registers any included packages, 
                      as they were currently in GOPATH/src
  
  unregister          removes a package from the registry
  
  diff                Show the difference in the GDFs between the local source
                      dir and its counterpart in the registry.
  
  lint                Check if the given package respects the recommendations
                      for a package maintainer as given by the GDF.
                      Please keep in mind that not all recommendations can be
                      automatically checked. This is highly WIP.
  
  init                (Re)initialize the registry for the whole GOPATH and
                      check for incompatibilities in exports between the 
                      packages in GOPATH/src. 
                      WARNING: this erases the former compatibility
                      informations in the registry and the checksums of the 
                      init functions that used to be compatible.

  init-functions      show the content of the init functions of the package, 
                      based on the local source dir
  
  check               checks the integrity of the whole GOPATH while based on
                      its registry.  

  registry-cleanup    removes orphaned packages in the registry that do not 
                      exist in GOPATH/src any longer  

  gopath-cleanup      removes orphaned temporary GOPATHs for tentative dep gets  

  backups-cleanup     removes all backups of dep get in the current GOPATH

  dump                dump the GDFs from the registry, sorted by package name
`
