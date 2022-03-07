# req

req is an opinionated HTTP scripting language. It is designed for easily making
HTTP requests, and working with their responses. Below is an example that calls
out to the GitHub API and displays the user making the call,

    $ cat gh.req
    Stderr = open "/dev/stderr";

    Endpoint = "https://api.github.com";
    Token = env "GH_TOKEN";

    if $Token == "" {
        writeln $Stderr "GH_TOKEN not set";
        exit 1;
    }

    Headers = (
        Authorization: "Bearer $(Token)",
    );

    Resp = GET "$(Endpoint)/user" $Headers -> send;

    match $Resp.StatusCode {
        200 -> {
            User = decode json $Resp.Body;

            writeln _ "Hello $(User["login"])";
        }
        _   -> {
            writeln $Stderr "Unexpected response: $(Resp.Status)";
            exit 1;
        }
    }
    $ GH_TOKEN=1a2b3c4d5ef req gh.req

This language hopes to fill in a gap when it comes to writing scripts for
working with an HTTP service. Typically, you have a choice between a shell
script that utilizes cURL, or a programmaing language and any HTTP library
that may come with it.

The cURL approach can work, for simple one off requests, but when you want to
do something more with the response you're left with having to munge that data
with jq, grep, sed, or awk (or all of the above). Using a programming language
gives you more control, but can be more cumbersome as you have far many more
knobs to turn.

req provides a middleground between the two. A limited syntax, with builtin
commands for working with any data you want to send/received. For more details
on how to start working with req, then refer to the [documentation](docs), or
you can dive right in by looking over the [examples](examples).
