# Control flow

Control flow in req is managed via `if`, `else`, `match`, `for`, `break`, and
`continue`.

* [If statements](#if-statements)
* [Match statements](#match-statements)
* [For loops](#for-loops)
* [Break and continue](#break-and-continue)
* [Logical operators](#logical-operators)
* [Equality operators](#equality-operators)

## If statements

`if` allows you to conditionally execute blocks of code. This takes a condition
that will either evaluate to `true` or `false`, followed by a block of code
wrapped in a pair of `{ }`,

    if true {
        # Do something
    }

`else` can be defined following an `if` statement to have another block of code
executed should the initial condition evaluate to `false`,

    if true {
        # Do something
    } else {
        # Do something else
    }

`else` and `if ` can be combined for more explicit control flow,

    if $Resp.StatusCode >= 200 and $Resp.StatusCode < 300 {
        writeln _ "Everything is good";
    } else if $Resp.StatusCode >= 400 and $Resp.StatusCode < 500 {
        writeln _ "We messed up";
    } else {
        writeln _ "They messed up";
    }

## Match statements

`match` allows you to conditionally execute blocks of code, similar to an `if`.
This takes a single condition, which should evaluate to a literal,

    match $Resp.StatusCode {
        200 -> writeln _ "Everything is good";
        400 -> writeln _ "We messed up";
        500 -> writeln _ "They messed up";
        _   -> {
            writeln _ "Unhandled status code";
            exit 1;
        }
    }

this differs from an `if-else` chain in how it is implemented, under the hood
this used a jump table to conditionally execute the blocks of code. You may
prefer to use a `match` statement in place of an `if-else` chain if you are
checking a value against multiple constants.

The condition must evaluate to a literal, and can only be matched against any
other literal (bool, string, or numeric). The `_` identifier in the `match`
statement is the default block of code to execute should none of the conditions
match.

## For loops

`for` allows you to execute a block of code multiple times depending on a
given condition. `for` loops can be defined with no condition for an infite
loop,

    for {
        # Executes indefinitely unless break is used
    }

a single condition can also be given to a `for` loop,

    for $Condition {
        if $Condition == false {
            break;
        }
    }

typical `for` loops can also be defined whereby there is an initializer, a
condition, and a post, for example to read through all the lines a file,

    F = open "events.json";

    for Line = readln $F; $Line != ""; Line = readln $F {
        # Do something with $Line
    }

the `range` keyword can be used to iterate over arrays and objects,

    Arr = [1, 2, 3, 4];

    for I, N = range $Arr {
        # Do something with the index and value
    }

    Obj = (
        String: "string",
        Array: [1, 2, 3, 4],
        Bool: true,
    );

    for K, V = range $Obj {
        # Objects are iterated over in the order they are created, so String
        # would be the first key, then Array, then Bool in this example.
    }

the `range` keyword produces two values when used in a `for` loop condition,
the key and the value of the iterable. An `_` identifier can be used ignore
one of the returned values. Single assignments to `range` are also valid,

    for I = range [1, 2, 3, 4] {
        # $I would be the index number for an array
    }

    for K = range (One: "", Two: "", Three: "") {
        # $K would be the key of the object
    }

## Break and Continue

`break` and `continue` can be used to control the flow of a `for` loop. `break`
would break out of the loop and stop subsequent execution. `continue` would stop
the current execution, and move on to the next iteration of the loop,

    for I = range [1, 2, 3, 4] {
        if $I == 1 {
            continue;
        }
        writeln _ $I;
    }

## Logical operators

Logical operators in req are used to evaluate two expression on either side of
an operator. In req, these are `and`, `or`, and `in`.

The `and` operator will evaluate to `true` if both sides of the operator
evaluate to `true`,

    1 == 1 and 1 != 0

The `or` operator will evaluate to `true` if either sides of the operator
evaluate to `true`,

    1 == 2 or 1 == 1

The `in` operator will evaluate to `true` if the left hand side can be found in
the right hand side, and the right hand side is either an array or object. When
the `in` operator is performed on an array it will evaluate to `true` if the
item exists int he array. When the `in` operator is performed on an object it
will evaluated to `true` if the key exists in the object,

    2 in [1, 2, 3, 4] # true
    "Key" in (Key: "value") # true
    "Foo" in (Key: "value") # false

## Equality operators

Equelity operators in req are used to evaluate two expressions on either side of
an operator to determine their equality. In req, these are,

    ==    equals
    !=    not equals
    <     less than
    <=    less than or equal
    >     greater than
    >=    greater than or equal

only expressions of the same type can be compared in this way. The following
values cannot be compared,

* [File](values.md#file)
* [FormData](values.md#formdata)
* [Request](values.md#request)
* [Response](values.md#response)
* [Stream](values.md#stream)

bools, objects, and arrays can only have the equals and not equals operators
used for comparison.
