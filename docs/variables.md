# Variables

Variables in req are defined with an identifier on the left hand side, and a
value on the right. This [value](values.md) can either be a literal, or an
evaluated value from a [command](commands.md),

    S = "string";
    I = 10;
    A = [1, 2, 3, 4];

variables can be defined on the same line too, with a comma separating each
identifier, and expression,

    S, I, A = "string", 10, [1, 2, 3, 4];

variables are referenced with `$` for when you want to use them elsewhere, as
a value to another variable or an argument to a [command](commands.md).

    S = "Hello world";
    Base64 = encode base64 $S;

Variables defined in a block will exist until the end of that block.

    if true {
        V = "block";
    }
    writeln _ $V; # results in an error, undefined: V

>**Note:** This is not the case in the REPL currently.
