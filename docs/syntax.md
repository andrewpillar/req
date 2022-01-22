# Syntax

A req script is a plain text file with `.req` suffixed to the name, Each script
file is a list of statements, ending with a semicolon (`;`).

* [Comments](#comments)
* [Keywords](#keywords)
* [Identifiers](#identifiers)

## Comments

Comments in req start with `#` and end with a newline.

    # Full-line comment.
    if true { } # Another comment.

## Keywords

Listed below are the keywords of req that cannot be used as identifiers,

    break
    continue
    if
    else
    for
    match
    range
    in
    and
    or

## Identifiers

Identifiers start with a letter or underscore and can contain letters, numbers,
underscores, and hyphens,

    contentType
    ContentType
    Content-Type
    Content_Type
    _Content_Type
