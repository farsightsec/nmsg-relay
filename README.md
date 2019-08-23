# NMSG uploader for Farsight SIE

A lightweight client which reads NMSG input from a datagram socket
and submits it to the SIE.

## Synopsis

    nmsg-relay -channel <channel-number>
                  [ -config /path/to/config.yaml ]
                  [ -input <address>/<portrange> ]
                  [ -apikey [key | /path/to/keyfile] ]
                  [ -heartbeat <interval> ]
                  [ -retry <interval> ]
                  [ -flush <interval> ]
                  [ -stats_interval <interval> ]
                  wss://server-1/session/nmsg-upload
                  [ wss://server-2/session/nmsg-upload ]

    nmsg-relay -version  # prints version and exits

## Building

The latest version can be built with

    go get github.com/farsightsec/nmsg-relay

And installed with:

    go install github.com/farsightsec/nmsg-relay

This package also has no non-go dependencies, so cross compilation
is supported with:

    env GOOS=<target> GOARCH=<target-arch> \
        go build github.com/farsightsec/nmsg-relay

## Configuration

Configuration of `nmsg-relay` can be stored in the environment, an optional
config file or specified on the command line. Environment values are overridden
by configuration file values. Configuration file values are overridden by
command line values.

The location of a configuration file is given with the `NMSG_RELAY_CONFIG`
environment variable or the `-config` command line option. There is no
default configuration file.

### Channel

The destination channel number for the nmsg data is a required parameter. It
can be specified in the environment variable `NMSG_RELAY_CHANNEL`, in the
config file with:

        channel: 42

or on the command line with:

        -channel=42

### Input

`nmsg-relay` reads NMSG data from one or more UDP sockets. The address or addresses
`nmsg-relay` takes input from is a required parameter.

A single socket address can be specified as `<ip>/<port>`, e.g.: "127.0.0.1/5053".
A range of ports on a single address can be specified as `<ip>/<lowport>..<highport>`,
e.g.:  "127.0.0.1/5050..5053".

The socket address specification can be given in the environment variable
`NMSG_RELAY_INPUT`, in the configuration file with:

        input: 127.0.0.1/5053

or on the command line with:

        -input=127.0.0.1/5053

### Heartbeat and Retry

The server connections maintained by `nmsg-relay` send periodic heartbeats
to instruct the server to keep the connection open. If the server connection
drops, `nmsg-relay` attempts to reconnect after a given `retry` interval.

These can be specified in the environment variables `NMSG_RELAY_HEARTBEAT`
and `NMSG_RELAY_RETRY`, in the config file with:

        heartbeat: 10s
        retry: 1s

or on the command line with:

        -heartbeat=10s -retry=1s

The interval is specified in the syntax supported by
(time.ParseDuration)[https://godoc.org/time#ParseDuration]. Both default to 30s.

### Flush Interval

`nmsg-relay` will attempt to combine multiple nmsg payloads into larger
containers. The flush interval provides a maximum time data will be buffered.
It can be specified in the environment variable `NMSG_RELAY_FLUSH`, the 
config file with:

        flush: 400ms

or on the command line with:

        -flush=400ms

The default is 500ms.

### API Key

`nmsg-relay` authenticates itself to the server with an API key. This can
be specified in the environment variable `NMSG_RELAY_APIKEY`, the config
file with:

        api_key: <key>

or:

        api_key: /path/to/keyfile

or on the command line with:

        -apikey=<key-or-file>

The API Key is a required parameter.

### Servers

The remainder of the command line arguments to `nmsg-relay` are treated as
a list of server URLs. If none are specified on the command line, the values
from the environment variable `NMSG_RELAY_SERVERS` or config file (if any)
are used. Servers can be specified in the config file with:

        servers:
            - wss://server-1-hostname/session/<name>
            - wss://server-2-hostname/session/<name>

At least one server must be specified. If the `/session/` path is not given,
it defaults to `/session/nmsg-relay-upload`.
