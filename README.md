# go-authproxy

This is a proxy which can

- require the client use basic authentication
- require the client go through a captive portal

# Installation

```sh
$ go get -u github.com/jacobdufault/go-authproxy
$ go-authproxy -h # prints out help/usage
```

# Examples

Force chrome to show basic authentication proxy connection UI:

```sh
# terminal A
$ go-authproxy -basic-auth user:pass # interrupt (e.g. <c-c>) to shutdown
# terminal B
$ chrome --proxy-server="127.0.0.1:8080"
```

## More examples

```sh
# Require a captive portal
$ go-authproxy -captive-portal

# Require basic authentication
$ go-authproxy -basic-auth user:pass

# Require both
$ go-authproxy -basic-auth user:pass -captive-portal

# Basic auth on port 8080
$ go-authproxy -basic-auth -port 8080
```

## Captive Portal Visibility

The captive portal is shown when a file called `dismiss-captive-portal` is
not present on disk in the current working directory. On startup,
go-authproxy automatically deletes this file. It can be manually deleted at
any time and the captive portal will be re-displayed.
