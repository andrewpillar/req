# req

req is an HTTP scripting language.

Below is an example that calls out to the GitHub API and displays the user
making the call,

    Stdout = open "/dev/stdout";
    Stderr = open "/dev/stderr";

    Endpoint = "https://api.github.com";
    Token = env "GH_TOKEN";

    Headers = {
        Authorization: "Bearer {$Token"},
    };

    Resp = GET "{$Endpoint}/user" $Headers -> send;

    match $Resp.StatusCode {
        200 -> print $Resp.Body;
        _   -> {
            print "Failed to send request" $Stderr;
            exit 1;
        }
    }
