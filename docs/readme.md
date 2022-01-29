# Documentation

req is a simple HTTP scripting language. It can either be run as a script or
the REPL can be used as playground. If starting out with req, it is suggested
that you use the REPL to get to grips with the language. The REPL is limited in
its capabilities however.

* [Installation](#installation)
* [Running](#running)
* [Conventions](#conventions)
  * [Formatting](#formatting)
  * [Comments](#comments)
  * [Names](#names)
* [Guide](#guide)

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

## Conventions

The following conventions should be followed when writing a req script. These
are meant to aid in the readability and consistency of the scripts that are
written.

### Formatting

Each req script should start with two blank lines at the top. This is not needed
however, if you intend on having an opening comment, in which case this goes
starts from the first line. If using a shebang, then the following two lines
should be blank,

    #!/usr/bin/env req
    
    
    # Beginning of script...

Use tabs for indentation, and spaces for alignment, for example,

    Headers = (
        Accept:        "application/json",
        Authorization: "Bearer $(Token)",
    );

When defining an object use trailing commas. When defining an array only use
trailing commas if the items are on a newline, for example,

    Arr = [1, 2, 3, 4]; # No trailing comma, all items on same line

    Arr = [
        1, 2, 3, 4, # Trailing comma, items start on a newline
    ];

    # Trailing comma, items each start on a newline
    Arr = [
        1,
        2,
        3,
        4,
    ];

There is no limit placed on the length of a line in a req script, though it is
suggested to aim to keep them between 80 - 100 characters in length. Exceeding
this length is fine when the circumstance necessitates it, this is something
that would only be known when it happens.

### Comments

Comments in req a prefixed with a `#`. It is typically preferred to have
comments on individual lines. Having a comment on the same line as an
expression is fine only if the length of the line can afford it.

### Names

Names in req should follow PascalCase. This is not mandated by the parsing of
a req script, but is suggested so at a glance you can delineate between a name
and a command,

    Obj = (
        Key: "value", # Use PascalCase for object keys.
    );

When defining an object that will be used as request headers, ensure the keys
are capitalized where necessary, and `-` separated where necessary. Even though
HTTP headers are case insensitive, it is recommended you adhere to the following
format for consistency across the rest of a script,

    Headers = (
        Content-Type: "application/json",
        User-Agent:   "my req script",
    );

## Guide

* [Syntax](syntax.md)
* [Values](values.md)
* [Control flow](control-flow.md)
* [Variables](variables.md)
* [Commands](commands.md)

[0]: https://go.dev
