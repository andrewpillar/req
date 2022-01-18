# req

req is an HTTP scripting language.

Below is an example that calls out to the GitHub API and displays the user
making the call,

    Stderr = open "/dev/stderr";

    Endpoint = "https://api.github.com";
    Token = env "GH_TOKEN";

    if $Token == "" {
        fprint $Stderr "GH_TOKEN not set";
        exit 1;
    }

    Headers = (
        Authorization: "Bearer {$Token}",
    );

    Resp = GET "{$Endpoint}/user" $Headers -> send;

    match $Resp.StatusCode {
        200 -> {
            User = decode json $Resp.Body;

            print "Hello {$User["login"]}";
        }
        _   -> {
            fprint $Stderr "Unexpected response:" $Resp.Status;
            exit 1;
        }
    }

this would then be run like so, as long as the given file is saved with the
`*.req` suffix and is in the current directory from which req is invoked, then
it will be evaluated.

    $ GH_TOKEN=<token> req
