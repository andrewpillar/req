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

A single req script can be executed by passing it to the `req` binary,

    $ req script.req

if no arguments are given to `req`, then the REPL will be opened up. This can
be used as a playground during the writing of req scripts,

    $ req
    > S = "string"
    > $S
    "string"
    >

* [Syntax](syntax.md)
* [Values](values.md)
* [Control flow](control-flow.md)
* [Variables](variables.md)
* [Commands](commands.md)

[0]: https://go.dev
