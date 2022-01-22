# Values

Values are the builtin types of req. These can either be created via literals,
or returned as a result of a [command](commands.md).

* [Bool](#bool)
* [String](#string)
* [Number](#number)
* [Array](#array)
* [Object](#object)
* [File](#file)
* [FormData](#formdata)
* [Request](#request)
* [Response](#response)
* [Stream](#stream)
* [Zero](#zero)

## Bool

A bool represents either `true` or `false`.

## String

A string is a sequence of bytes wrapped between a pair of `"`. These support
the escape sequences `\t`, `\r`, and `\n`. Strings also support interpolation
via `$( )`,

    "Hello $(Name)"
    "$(Obj["Key"])" # Double quotes do not need escaping during interpolation

## Number

A number is a numeric value. This can either be an integer for a float. As of
now req does not support working with negative numeric values, or any other
numeric value that is not of base 10.

## Array

An array is a list of values. Arrays defined by the user can only contain one
of the same type. The [zero](#zero) value is returned if out of bounds array
access is made,

    Arr = [1, 2, 3, 4];
    $Arr[5] # Does not error, results in zero value

## Object

An object is a list of key-value pairs, wrapped in a pair of `( )`. Keys defined
in an object, are defined as identifiers,

    Obj = (
        Authorization: "Bearer 1234",
    );

if an non-existent key is accessed in an object, then the [zero](#zero) vlaue is
returned.

    $Obj["Key"] # Does not error, results in zero value

## File

A file represents a file that has been accessed via the [open](commands.md#open)
command. A file can be read from and written to. A file can be used as a
[Stream](#stream).

## FormData

[Form data][0] represents the form data being sent as the body of an HTTP
request. This is created via the [encode form-data](commands.md#encoding) command
and would be typically used for doing file uploads. This has the following
properties on it,

* `Content-Type` - `string` - The `Content-Type` header for the form-data. This
would be set in the header of a request being sent

* `Data` - `Stream` - The file data of the form-data

## Request

Request represents an HTTP request being sent. This is created via one of the
[request](commands.md#requests) commands. This has the following properties on
it,

* `Method` - `string` - The HTTP method of the request

* `URL` - `string` - The URL the request will be sent to

* `Header` - `Object` - The headers set on the request

* `Body` - `Stream` - The body of the request

## Response

Response represents an HTTP respons that has been received. This is created via
the [send](commands.md#send) command. This has the following properties on it,

* `Status` - `string` - The HTTP status of the response, such as `200 OK`

* `StatusCode` - `int` - The status code of the response

* `Header` - `Object` - The headers set on the response

* `Body` - `Stream` - The body of the response

## Stream

Stream represents a stream of read-only data. This will either be a buffer of
data that exits in memory, or from another source such as an opened file.

## Zero

Zero represents a zero value. A zero value is created when an invalid access
to an array or object is made. A zero value will evaluate to `true` when
compared to another zero value of any type,

    Arr = [];
    Zero = $Arr[1];

    $Zero == "" # true
    $Zero == 0 # true
    $Zero == false # true
    $Zero == [] # true
    $Zero == () # true

[0]: https://developer.mozilla.org/en-US/docs/Learn/Forms/Sending_and_retrieving_form_data
