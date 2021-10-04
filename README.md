[![Build Status](https://github.com/azazeal/vasc/actions/workflows/build.yml/badge.svg)](https://github.com/azazeal/vasc/actions/workflows/build.yml)
[![Coverage Report](https://coveralls.io/repos/github/azazeal/vasc/badge.svg?branch=master)](https://coveralls.io/github/azazeal/vasc?branch=master)
[![Go Reference](https://pkg.go.dev/badge/github.com/vasc.svg)](https://pkg.go.dev/github.com/azazeal/vasc)

# vasc

Package vasc implements a client for Varnish Administrative Socket interfaces.

# Disclaimer

`vasc` hasn't been used in the wild; it may be broken. Additionally, I offer
no compatibility guarantees at this time.

## Simple usage

Assuming you have a Varnish instance

1. on host `varnish`
2. listening on port `10000`
3. with its secret being `mySuperStronkVarnishSecret`

then in order to retrieve the instance's `pid` information, you'd want to use
something like this:

```go
package main

import (
	"log"
	"time"

	"github.com/azazeal/vasc"
)

func main() {
	cfg := vasc.Config{
		Secret:       "mySuperStronkVarnishSecret",
		ReadTimeout:  time.Minute, // Maximum amount of time to allow for reads
		WriteTimeout: time.Minute, // Maximum amount of time to allow for writes
	}

	client, err := vasc.Dial("tcp", "localhost:10000", cfg)
	if err != nil {
		log.Fatalf("failed dialing varnish: %v", err)
	}
	defer client.Close()

	code, data, err := client.Execute(nil, "pid", "-j")
	switch {
	case err != nil:
		log.Fatalf("failed executing: %v", err)
	case code == vasc.StatusClose:
		log.Fatal("varnish closed the connection!")
	case code != vasc.StatusOK:
		log.Fatalf("request failed with status: %d", code)
	}

	var pid struct {
		Master int `json:"master"`
		Worker int `json:"worker"`
	}

	if err := vasc.UnmarshalJSONResponse(data, &pid); err != nil {
		log.Fatalf("failed unmarshaling JSON response: %v", err)
	}

	log.Printf("master pid: %d", pid.Master)
	log.Printf("worker pid: %d", pid.Worker)
}
```

The full list of commands that Varnish supports may be found
[here](https://varnish-cache.org/docs/7.0/reference/varnish-cli.html#commands).
