# Values

Values are the builtin types of req. These can either be created via literals,
or returned as a result of a [command](commands.md).

* [bool](#bool)
* [string](#string)
* [number](#number)
* [array](#array)
* [object](#object)
* [file](#file)
* [form-data](#form-data)
* [cookie](#cookie)
* [request](#request)
* [response](#response)
* [stream](#stream)
* [zero](#zero)

## bool

A bool represents either `true` or `false`.

## string

A string is a sequence of bytes. String literals are defined by wrapping the
bytes between a pair of double quotes (`"`),

    "String"

these are immutable, in that, you cannot modify individual parts of a string,

    S = "String";
    S[2] = ""; # Not allowed

Strings support the `\t`, `\r`, and `\n` escape sequences. Strings also support
interpolation, whereby you can refer to previously defined variables within a
string literal. This is achieved with a dollar (`$`) followed by a pair of
parentheses, (`( )`),

    Arr = [1, 2, 3, 4];
    S = "Arr = $(Arr)"; # Arr = [1 2 3 4]

when using interpolation to refer to a key in an object, you do not need to
escape the double quotes within the parentheses,

    Obj = (Key: "value");
    S = "Object key = $(Obj["Key"])"; # Object key = value

## number

A number is a numeric value. This can either be an integer for a float. As of
now req does not support working with negative numeric values, or any other
numeric value that is not of base 10.

    10
    10.25

## array

An array is a list of values. Arrays defined by the user can only contain one
of the same type. The [zero](#zero) value is returned if out of bounds array
access is made,

    Arr = [1, 2, 3, 4];
    Z = $Arr[5]; # Does not error, results in zero value

arrays are mutable, in that, you can modify the individual items in an array,
as long as the new value is of the same type,

    Arr = [1, 2, 3, 4];
    Arr[0] = 0;
    Arr[1] = "2"; # Not allowed

## object

An object is a list of key-value pairs, wrapped in a pair of `( )`. Keys defined
in an object, are defined as identifiers,

    Obj = (
        Authorization: "Bearer 1234",
    );

if an non-existent key is accessed in an object, then the [zero](#zero) vlaue is
returned.

    $Obj["Key"] # Does not error, results in zero value

much like an array, individual items can be modified, as long as the new value
is of the same type,

    Obj = (Key: "value");
    Obj["Key"] = "new value";
    Obj["Key"] = 2; # Not allowed
    Obj["Page"] = 2; # Allowed, creates a new key

## file

A file represents an open file descriptior. The file value is returned from the
[open](commands.md#open) command. A file can be read from and written to. A file
can be used as a [stream](#stream).

## form-data

form-data represents the encoded data for a `multipart/form-data` of a form.
This is created via the [encode form-data](commands.md#form-data) command and
would typically be used for creating request bodies that would adhere to the
[RFC 2388][RFC-2388] format when you want to submit `multipart/form-data` to an
endpoint.

FormData is an entity with the following properties on it,

**`Content-Type`** - [string](#string) - The `Content-Type` header for the
encoded data. This would be set in the header of a request.

**`Data`** - [stream](#stream) - The raw bytes of encoded data. This would be
used as the body of the request.

## cookie

Cookie represents an HTTP cookie that can be sent in a request, or is sent in
an HTTP response. This is created via the [cookie](commands.md#cookie) command,
and can be retrived from a [response](#response) via the `Cookie` field.

Cookie is an entity with the following properties on it,

**`Name`** - [string](#string) - The name of the cookie.

**`Value`** - [string](#string) - The value of the cookie.

**`Path`** - [string](#string) - The path of the cookie.

**`Domain`** - [string](#string) - The domain of the cookie.

**`Expires`** - time - When the cookie expires.

**`MaxAge`** - duration - The max age of the cookie.

**`Secure`** - [bool](#bool) - Whether or not the cookie is secure.

**`HttpOnly`** - [bool](#bool) - Whether or not the cookie is HTTP only.

**`SameSite`** - [string](#string) - How the cookie should be restricted.

## request

Request represents an HTTP request. This is created via one of the
[request](commands.md#request) commands.

Request is an entity with the following properties on it,

**`Method`** - [string](#string) - The HTTP method of the request.

**`URL`** - [string](#string) - The URL the request will be sent to.

**`Header`** - [object](#object) - The headers set on the request. This can
only be set at time of request creation and not after.

**`Body`** - [stream](#stream) - The raw bytes of the request body.

## response

Response represents an HTTP response that has been received. This is created
via the [send](commands.md#send) command.

Response is an entity with the following properties on it,

**`Status`** - [string](#string) - The HTTP status of the response, such as
`200 OK`.

**`StatusCode`** - [int](#number) - The status code of the response.

**`Cookie`** - [object](#object) - The [cookies](#cookie) sent in the response.

**`Header`** - [object](#object) - The headers set on the response.

**`Body`** - [stream](#stream) - The raw bytes of the response body.

## stream

Stream represents a stream of read-only data. This will either be a buffer of
data that exits in memory, or from another source such as an opened file.

## zero

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

[RFC-2388]: https://datatracker.ietf.org/doc/html/rfc2388 
