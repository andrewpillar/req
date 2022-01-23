# Documentation

req is a simple HTTP scripting language. It can either be run as a script or
the REPL can be used as playground. If starting out with req, it is suggested
that you use the REPL to get to grips with the language. The REPL is limited in
its capabilities however.

* [Installation](#installation)
* [Running](#running)

## Installation

>**Note:** The following documentation assumes that you already have some
familiarity with programming.

req can be installed from source, doing so requires [Go][0]. First, clone the
repository,

    $ git clone https://github.com/andrewpillar/req

then change into it and run the `make.sh` script,

    $ cd req
    $ ./make.sh

this will produce a `req` binary in the `bin` directory. Simply add this to your
`PATH`.

## Running

When not arguments are given to `req` then the REPL will be opened up. This can
be used a playground during the writing of req scripts,

    $ req
    > S = "string"
    > $S
    "string"
    >

each argument given to req will be the path to a script or scripts to be
executed. If the path is a directory then all of the req scripts in the top
level of that directory are executed,

    # Execute all scripts in the current directory
    $ req .

    $ req 0.req 1.req 2.req

each script executed is executed concurrently.

* [Syntax](syntax.md)
* [Values](values.md)
* [Control flow](control-flow.md)
* [Variables](variables.md)
* [Commands](commands.md)

[0]: https://go.dev
