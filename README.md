dep
===

__WARNING: This is currently a draft and WIP. The concept and the tools are in pre-alpha state and not meant to be used in production. You may break your installation.__

Manages Go package dependencies with the help of the 
[Go Dependency Format](http://github.com/go-dep/gdf) (GDF) 
that is based on the exports of a package.

The idea is described [here](https://docs.google.com/document/d/1hN7OP4QjfsasWvKSvm3NdjW1-3tKdFkmCeSI3uaT6wo).

dep is a commandline tool that helps you discover problems with 
package dependencies before they affect your development environment.

Help on the subcommands is available via

    dep -h

Prerequisites:

  - A functional GOPATH. If GOPATH is set to multiple paths, 
    separated by ':', only the first one is considered
  - The environment variable DEP_TMP set to a temporary directory.
  - Make sure no package you have uses relative import paths, since 
    they are not supported and a bad habit anyway.
  - You need a working github.com/mattn/go-sqlite3 package.

WARNING: Currently dep is not production ready and may break in 
several ways. So by now it should only be used for testing purposes. 
Setup a seperate GOPATH to do the testing of dep without affecting 
your real packages.
    
To get the most out of dep, please use the following workflow:

Initialization
--------------

If you use dep for the first time in a GOPATH, run 
        
    dep init

this will initialize the registry and register all packages in 
GOPATH/src/dep.db.

You might get errors if they are any packages that have dependencies 
that are not met. 

You might either fix them, or (not recommended) put them in a 
GOPATH/src/.depignore file (each line a package path).
If they are ignored, dep is not able to track their dependencies 
and the dependencies of other packages importing them.

So continue to fix broken packages until you get no more errors 
from 'dep init'.

Now you will develop something. It is a good idea to do a

    dep register

inside the directory you are developing every time you do a commit or 
a go install.

Before installing / updating a package as described below, you should 
always run

    dep check

which checks, if your registry is up to date with the packages in your 
GOPATH and if there are conflicts. 

With a disfunctional registry there is nothing, dep could do to save 
you from breakage.

To install or update a package use

    dep get [package]

instead of the usual go get command. This does several things:

  - It does a tentative go get for the package and its imported 
    packages, checking out revisions specified in a dep-rev.json 
    if there is any in the package directory
  - It does check if the dependencies for that package and its 
    imported packages changed - with the help of the GDF - and if 
    there were any conflicts.
  - It there were conflicts, you get a detailed message and can 
    explore the issue in the temporary GOPATH where the tentative 
    'go get' did take place.
  - If there were no conflicts, the packages will be installed in 
    your current GOPATH.

If you working on a package within your team or you want to make your 
package accessible to others, it is a good idea to run the command

    dep track

inside the directory of your package, so that all revisions of the 
packages, your package imports and of their imported packages and so 
on are tracked in a dep-rev.json file inside your package directory. 

You should put this file under version control, so that others using
the dep tool can be sure to get a working dependency chain, if they 
dep get your package.
Or if there is a conflict they at least get informed before their 
environment is harmed.

Just to make it clear: The dep tool is able to find incompatibilities 
between even those packages that don't know about dep and have no 
dep-rev.json file.

However without the dep-rev.json it might be that even the first 
installation of package fails, if imported packages went incompatible. 
It will take some effort then to find out, which revision was the 
last that did work and that is the information tracked by the 
dep-rev.json file.

As a package developer you may minimize breakage and false positives 
of the dep tool, if you act according to the  
[package developer rules](https://github.com/go-dep/gdf/wiki/Recommendations-for-go-package-developers).

The command 
    
    dep lint

checks some of this rules.

To get an idea of what is stored a dependency information inside 
the registry, run

    dep gdf

inside the package directory.

It is a good idea, to check with 

    dep diff

what has changed in your package's gdf since the last time you did 
run dep register.

Then you will see what will break for users of your package if they 
do an update of your package.

If you removed a package, you need to tell dep of the removal with

    dep unregister

Typical Workflow
----------------

After you did set up your registry properly (with dep init) you should 
use the following workflow.

- Installation of new package or update of a package:
    - run 'dep check' to see if your environment is consistent, 
      if not fix it first and then start again
    - run 'dep get package-path'
    - if it went well, run 'dep check' again, to see which packages 
      were changed to what
    - run 'dep register package-path' for each package that changed
    - run 'dep check' again until your environment is ok again
    - run your tests to check if everything still works 
      (dep makes backups in the same directory as the original repo)
    - run 'go install' for the packages you want to have installed
    - run 'dep backups-cleanup' to remove all backup folders if you 
      don't need them anymore

- Fixing errors during the run of 'dep get':
    - You will be given a temporary GOPATH where all packages that 
      needed to be installed or updated are installed. First you will 
      want to know, if they are consistent, so set the GOPATH 
      environment variable to this folder and then run 'dep check' 
      to see problems.
    - If there are problems, you will have to decide what to do. 
      Maybe you could fix the packages and send them to upstream or 
      you decide to better do no update.
    - Don't forget to set your GOPATH back to the original one. 
      Do it now.
    - If the packages in the temporary GOPATH are fine, there will 
      be conflicts with your currently installed packages in the 
      original GOPATH. 'dep get' should have informed you exactly 
      about the conflicts. Now you may either fix your local packages 
      or decide to do no update. If you choose the latter, run
      'dep gopath-cleanup' to remove all temporary GOPATHs 
      and you are done.
    - If you want to fix the packages, first get the gdf for all of 
      them by running 'dep gdf pkgPath'. 
      The gdfs are in readable json format. 
      Then put them into a json file, say 'override.json' as part of 
      a json array. Now you can modify this file to match the exports 
      that are requested by the packages that are part of the 
      'dep get' process. Make sure that the package also works with
      the exports as defined in your file. Then you could try further 
      runs of 'dep get' with passing the override.json file as 
      parameter: 'dep -override=override.json get packagepath'. 
      This will take the gdfs given in the file overriding their 
      definition in the registry. If everything runs fine, the new 
      packages will be moved to your GOPATH and then you should make 
      sure that everything works as expected. After that you can run 
      'go install' and 'dep backups-cleanup'.

- Fixing errors reported by 'dep check':
    - 'dep check' does not report all errors, so you will have to run 
      it until no error is shown
    - If 'dep check' reports errors, this means that the registry is 
      not in sync with the packages in your GOPATH. 
      There might be several issues:

      - If a package is not inside the registry, register it with 
        'dep register' and run 'dep check' again. If you want to also 
        register the packages that are imported by the package run 
        'dep register-included <package-path>'

      - If a package is in the registry, but not in the GOPATH anymore 
        and if that is what you wanted, run 'dep unregister' and then 
        'dep check' again. If you want to remove all orphaned packages 
        from the registry, run 'dep registry-cleanup'.

      - If a package has changed and its old gdf is in the registry, 
        but you want to update it, run 'dep register <packagepath>' 
        and then 'dep check' again. Otherwise you might want fix it.
        You can get the changes between the current package and the 
        registry for only one package with 'dep diff <packagepath>'.

      - If a package is in conflict with another, you need to fix it 
        and then update the registry for the changed package with 
        'dep register <packagepath>'. Run 'dep check' again to see 
        if there are other issues.

      - If a package that is part of a larger repo is in conflict, 
        but you are not interested in it, you might put it into a 
        '.depignore' file to tell 'dep' to ignore it. You might need 
        to run 'dep unregister <packagepath>' additionally.
        However this is a dangerous operation, especially if you want 
        to use this package at a later time, but forget that you had 
        put it into .depignore.
        Packages in backup directories and directories with the name 
        'example' or 'examples' are always ignored.

      - If you messed up the registry, but know that your packages 
        have no conflicts, you might as well run 'dep init' to 
        regenerate the registry. This may take a while, depending on 
        the number of packages you have.
      
Test
----

Since the tests involve that a package changes its exports, 
they are in seperate repositories.
Currently there are the following test scenarios 
(you are invited to write more):

- [compatible changing symbols](http://github.com/go-dep/deptest_compatible)
- [incompatible changing symbols](http://github.com/go-dep/deptest_incompatible)
- [partial changing symbols (working)](https://github.com/go-dep/deptest_partial_working)
- [partial changing symbols (failing)](https://github.com/go-dep/deptest_partial_failing)
- [package breaking update of another package in same repo](https://github.com/go-dep/deptest_partial)
- [disappearing symbols](https://github.com/go-dep/deptest_missing)

To run them run ./runtests.sh

If you install the testing packages to work on the them, please be aware, that some tests (e.g. https://github.com/go-dep/deptest_missing) will break the consistency of your registry, i.e. when you run

    dep check

you will get an error. That is natural since the packages are meant to test the breakage and the dep tool is reporting these errors. 

However, to able to use the dep tool alongside with these packages, you will have to add them to a .depignore file inside your GOPATH. You should not do this for "normal" packages however, since they are now ignored by dep.


Issues or how you can help
--------------------------

-   only tested on linux, needs testers and code changes (mainly paths) 
    for MacOSX and Windows, but should be no big deal
    must be gained by evalution
-   improve documentation
-   improve code
-   more test cases
-   currently only tested with git. needs testing for bzr and hg.