dep
===

__WARNING: This is currently a draft and WIP. The concept and the tools are in pre-alpha state and not meant to be used in production. You may break your installation.__

Manages Go package dependencies with the help of the [Go Dependency Format](http://github.com/metakeule/gdf) (GDF) that is based on the exports of a package.

The idea is described [here](https://docs.google.com/document/d/1hN7OP4QjfsasWvKSvm3NdjW1-3tKdFkmCeSI3uaT6wo).

dep is a commandline tool that helps you discover problems with package dependencies before
the affect your development environment.

Help on the subcommands is available via

    dep -h

Prerequisites:

    - A functional GOPATH. If GOPATH is set to multiple paths, separated by ':', only the first one is considered
    - The environment variable DEP_TMP set to a temporary directory.
    - Make sure no package you have uses relative import paths, since they are not supported and a bad habit anyway.

WARNING: Currently dep is not production ready and may break in several ways.
So currently it should only be used for testing purposes. Setup a seperate GOPATH
to do the testing of dep without affecting your real packages.
    
To get the most out of dep, please use the following workflow:

Initialization
--------------

If you use dep for the first time in a GOPATH, run 
        
    dep init

this will initialize the registry and register all packages in GOPATH/src/dep.db.
You might get errors if they are any packages that have dependencies that are not met.
You might either fix them, or (not recommended) put them in a GOPATH/src/.depignore file.
If they are ignored, dep is not able to track their dependencies and the dependencies of other
packages importing them.

So continue to fix broken packages until you get no more errors from dep init.

Now you will develop something. It is a good idea to do a

    dep register

inside the directory you are developing every time you do a commit or a go install.

Before installing / updating a package as described below, you should always run

    dep check

which checks, if your registry is up to date with the packages in your GOPATH and if there
are conflicts. With a disfunctional registry there is nothing, dep could do to save you from
breakage.

To install or update a package use

    dep get [package]

instead of the usual go get command. This does several things:

    - It does a tentative go get for the package and its imported packages, checking out revisions specified in a dep-rev.json if there is any in the package directory
    - It does check, if the dependencies for that package and its imported packages changed, as define by the GDF and if there were any conflicts.
    - It there were conflicts, you get a detailed message and can explore the issue in the temporary GOPATH where the tentative go get did take place.
    - If there were no conflicts, the packages will be installed in your current GOPATH.

If you working on a package within your team or you want to make your package accessible to
others, it is a good idea to run the command

    dep track

inside the directory of your package, so that all revisions of the packages, your package imports and of their imported packages and so on are tracked in a dep-rev.json file inside
your package directory. You should put this file under version control, so that others using
the dep tool can be sure to get a working dependency chain, if they dep get your package.
Of if there is a conflict they at least get informed before their environment is harmed.

Just to make it clear: The dep tool is able to find incompatibilities between even those packages that don't know about dep and have no dep-rev.json file.

However without the dep-rev.json it might be that even the first installation of package fails, if imported packages went incompatible. It will take some effort then to find out, which
revision was the last that did work and that is the information tracked by the dep-rev.json file.

As a package developer you may minimize breakage and false positives of the dep tool, if you
act according to the  [package developer rules](https://github.com/metakeule/gdf/wiki/Recommendations-for-go-package-developers).

The command 
    
    dep lint

checks some of this rules.

To get an idea of what is stored a dependency information inside the registry, run

    dep gdf

inside the package directory.

It is a good idea, to check with 

    dep diff

what has changed in your package's gdf since the last time you did run dep register.
Then you will see what will break for users of your package if they do an update of your package.

If you removed a package, you need to tell dep of the removal with

    dep unregister

Test
----

Since the tests involve that a package changes its exports, they are in seperate repositories.
Currently there are the following test scenarios (you are invited to write more):

- [compatible changing symbols](http://github.com/metakeule/deptest_compatible)
- [incompatible changing symbols](http://github.com/metakeule/deptest_incompatible)
- [partial changing symbols](https://github.com/metakeule/deptest_partial)
- [disappearing symbols](https://github.com/metakeule/deptest_missing)

To run them: go get them and then run go test in the corresponding directory.
Warning some tests (e.g. https://github.com/metakeule/deptest_missing) might break
the consistency of your registry, i.e. when you run

    dep check

you will get an error. That is natural since the packages are meant to test the breakage
and the dep tool is indicating that. However, to able to use the dep tool alongside with
these packages, you will have to add them to a .depignore file inside your GOPATH.
You should not do this for "normal" packages however, since they are now ignored by dep.


Issues or how you can help
--------------------------

-   only tested on linux, needs testers and code changes (mainly paths) 
    for MacOSX and Windows, but should be no big deal
-   currently no solution for imports that are merged into the namespace (.)
-   currently no solution for getting the type of variables that must be gained by evalution
-   improve documentation
-   improve code
-   more test cases
-   currently only tested with git. needs testing for bzr and hg.