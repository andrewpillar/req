# Commands

* [Overview](#overview)
* [IO](#io)
  * [open](#open)
  * [read](#read)
  * [readln](#readln)
  * [sniff](#sniff)
  * [write](#write)
  * [writeln](#writeln)
* [General](#general)
  * [env](#env)
  * [uuid](#uuid)
  * [exit](#exit)
* [Encoding](#encoding)
  * [base64](#base64)
  * [form-data](#form-data)
  * [json](#json)
  * [url](#url)
* [Decoding](#decoding)
  * [base64](#base64-1)
  * [form-data](#form-data-1)
  * [json](#json-1)
  * [url](#url-1)
* [Requests](#requests)
  * [cookie](#cookie)
  * [send](#send)
  * [tls](#tls)


## Overview

Commands in req allow for the sending of requests, encoding/decoding of data
and working with streams of data. All commands in req are builtin to the
language, there are no user defined commands.

A command is invoked by specifying the name of the command and passing the
necessary arguments,

    encode base64 "Hello world";

the arguments are given to a command as a space separated list. Some commands
return values that can be assigned to variables,

    Enc = encode base64 "Hello world";

commands can be chained together with an arrow (`->`), this will take the
result of one command and pass it as the last argument to the subsequent
command,

    encode base64 "Hello world" -> decode base64;

this can make writing commands that require taking the results of previous
commands much easier,

    GET "https://example.com" -> send;

## IO

The following commands are used for basic IO operations, opening of files,
reading from streams and files, as well as writing to files.

### open

    open <string>

The open command takes a single argument that is the path to the file to open.
If the file fails to open then the script is terminated. If the given file does
not exist, then it is created. The returned [file](values.md#file) will be
opened for reading, writing, and appending.

    F = open "request.log";

### read

    read <stream>

The `read` command takes a single argument that is the [stream](values.md#stream)
to be read from. This will read the entire contents of the stream and return it
as a [string](values.md#string). If the given argument is an `_` identifier,
then it will read the entire contents of standard input,

    F = open "events.json";
    S = read $F;

    S = open "events.json" -> read;

    S = read _; # Read from standard input

### readln

    readln <stream>

The readln command takes a single argument that is the
[stream](values.md#stream) to be read from. This will read up to and including
the first newline character that it encounters from the given stream. If the
given argument is an `_` identifier, then it will read the entire contents of
standard input,

    F = open "events.json";
    S = readln $F;

    S = readln _;

### sniff

    sniff <stream>

The sniff command inspects the first 512 bytes of a [stream](values.md#stream)
and returns the MIME type for that stream. If no MIME type can be detected then
`application/octet-stream` is returned,

    F = open "image.jpg";

    sniff $F -> writeln _; # image/jpeg

### write

    write <stream> [values...]

The write command writes the given [values](values.md) to the given output
[stream](values.md#stream). If the given argument is an `_` identifier, then it
will be written to standard output,

    # This writes the contents of the file.
    open "events.json" -> write _;

    # This writes the verbatim contents of the request.
    GET "https://example.com" -> write _;

    # This writes the verbatim contents of the response.
    GET "https://example.com" -> send -> write _;

### writeln

    writeln <stream> [values...]

The writeln command is similar to the write command in how it functions. Only
it terminates everything written with a `\n` character.

## General

### env

    env <string>

The env command returns the environment variable with the given name. If no
environment variable is set, then an empty [string](values.md#string) is
returned,

    Token = env "GH_TOKEN";

### uuid

    uuid

The uuid command returns a new UUIDv4,

    Uuid = uuid;

### exit

    exit <int>

The exit command terminates script execution, and exits with the given code,

    exit 1;

## Encoding

The encoding family of commands can be used for encoding various
[values](values.md) into different data formats. Each of these commands operate
in a similar fashion, whereby the first argument to the `encode` command is an
identifier which is the sub-command to invoke.

### base64

    encode base64 <stream|string>

The `encode base64` command encodes the given value into base64. This returns
a [string](values.md#string) for the encoded results,

    Basic = encode base64 "admin:$(Password)";
    Enc = open "image.jpg" -> encode base64;

### form-data

    encode form-data <object>

The `encode form-data` command encodes the given value into a
[form-data](values.md#form-data) value that can be sent as a
[request](values.md#request) body,

    F = open "avatar.jpg";

    FormData = encode form-data (
        Name: "avatar",
        File: $F,
    );

    POST "https://example.com" (
        Content-Type: $FormData.Content-Type,
    ) $FormData.Data -> send;

### json

    encode json <array|object>

The `encode json` command encodes the given value into JSON. This returns the
JSON [string](values.md#string) for the encoded results,

    Obj = encode json (Username: "admin", Password: "secret");
    Arr = encode json ["foo", "bar", "zap"];

### url

    encode url <object>

The `encode url` command encodes the given value into a URL encoded
[string](values.md#string),

    # Would be encoded to Password=secret&Perms=read&Perms=write&Username=admin
    URL = encode url (
        Username: "admin",
        Password: "secret",
        Perms: ["read", "write"],
    );

## Decoding

The decoding family of commands act as the inverse of the Encoding family of
commands. Each of these commands will decode a data format into the native
[value](values.md).

### base64

    decode base64 <stream|string>

The `decode base64` command decodes the given value from the base64
representation. This returns a [stream](values.md#stream) of the decoded value,

    Enc = encode base64 "Hello world";
    Stream = decode base64 $Enc;

    encode base64 "Hello world" -> decode base64;

### form-data

    decode form-data <form-data>

The `decode form-data` command decodes the given value from the
[form-data](values.md#form-data) representation. This return an object of the
decoded value,

    Obj = encode form-data (
        File: open "avatar.jpg",
    ) -> decode form-data;

### json

    decode json <stream|string>

The `decode json` command decodes the given value from the JSON representation.
This returns either an [array](values.md#array), [object](values.md#object),
[number](values.md#number), [bool](values.md#bool), or
[string](values.md#string), depending on what is being decoded,

    decode json "[1, 2, 3, 4]" # [1 2 3 4]
    decode json "{\"title\": \"Scripting in req\"}" # (title:"Scripting in req")

### url

    decode url <string>

The `decode url` command decodes the given value from the URL encoded
representation. This returns an [object](values.md#object) with the decoded
values,

    # Becomes (page:10 category:Programming)
    decode url "page=10&category=Programming"

## Requests

    METHOD <string> [object] [stream|string]

Requests are created by using one of the following commands,

    HEAD
    OPTIONS
    GET
    PUT
    POST
    PATCH
    DELETE

the first argument to the command must be the URL that the
[request](values.md#request) is for. The second argument is an
[object](values.md#object) detailing the headers for the request, and the third
is the request body. The final two arguments are optional. The methods, `HEAD`,
`OPTIONS`, `GET`, and `DELETE` ignore the third argument,

    GET "https://example.com" (Accept: "application/json");

    Payload = open "payload.json";
    POST "https://example.com" (Content-Type: "application/json") $Payload;

### cookie

    cookie <object>

The `cookie` command creates a cookie that can be given to a request in the
`Cookie` header field. This takes an object that expects the following fields,

    Name     string
    Value    string
    Path     string
    Domain   string
    MaxAge   duration
    Secure   bool
    HttpOnly bool

this is then given to the `Cookie` field in the request header, for example,

    Cookie = cookie (Name: "session", Value: encode base64 "session_token");
    Req = GET "https://example.com" (Cookie: $Cookie);

this can also be given an array of cookies too,

    SessionCookie = cookie (Name: "session", Value: encode base64 "session_token");
    LoginCookie = cookie (Name: "login", Value: "logged_in");

    Cookies = [
        $SessionCookie,
        $LoginCookie,
    ];

    Req = GET "https://example.com" (Cookie: $Cookies);

### send

    send <request>

The `send` command sends the given [request](values.md#request). This returns a
[response](values.md#response),

    Req = GET "https://example.com";
    Resp = send $Req;

    Resp = GET "https://example.com" -> send;

### tls

    tls [string] [string] [string] <request>

The `tls` command configures the given [request](values.md#request) for
transport via TLS. If the only argument given to this command is the request
itself, then only the system certificate pool is used as the root CA. If the
first argument is given and is an `_` identifier, then the system certificate
pool is used as the root CA, otherwise a [string](values.md#string) is expected.
This string can either be a path to a file, or a directory containing multiple
certificates. If the second and third arguments are both given then these are
used the the paths to the certificate and key to use respectively.

    # Uses system certificate pool for root CA.
    GET "https://example.com" -> tls;

    # Uses the system certificate pool for root CA, and the given client.crt
    # and client.key files.
    GET "https://example.com" -> tls _ "client.crt" "client.key";

    # Uses the contents of the /etc/ssl/certs directory for the root CA, and
    # the given client.crt and client.key files.
    GET "https://example.com" -> tls "/etc/ssl/certs" "client.crt" "client.key";

    # Uses the contents of /etc/ssl/ca.crt for the root CA and the given
    # client.crt and client.key files.
    GET "https://example.com" -> tls "/etc/ssl/ca.crt" "client.crt" "client.key";

    # This can be chained with the send command like so for TLS transport.
    GET "https://example.com" -> tls -> send;
