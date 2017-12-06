# Fakeadog

[![Documentation](https://godoc.org/github.com/johnstcn/fakeadog?status.svg)](http://godoc.org/github.com/johnstcn/fakeadog)
[![Go Report Card](https://goreportcard.com/badge/github.com/johnstcn/fakeadog)](https://goreportcard.com/report/github.com/johnstcn/fakeadog)
[![Build Status](https://travis-ci.org/johnstcn/fakeadog.svg?branch=master)](https://travis-ci.org/johnstcn/fakeadog)


Inspired by [Lee Hambley's Ruby script](http://lee.hambley.name/2013/01/26/dirt-simple-statsd-server-for-local-development.html).

Fakeadog can be used as an aid for testing emitting DataDog metrics locally without having to install a full-blown DataDog client.

Usage: `fakeadog -host $HOST -port $PORT`

To install: ```go get -u github.com/johnstcn/fakeadog```

The program leverages the library `fakeadog/parser` for parsing DataDog events from raw UDP packets.

Example usage:
```
import "github.com/johnstcn/fakeadog/parser"

func main() {
    parser := parser.NewDataDogParser()
    payload := []byte{"myapp.frobble.count:1|c|#app:myapp,hostname:myhost"}
    metric, err := parser.Parse(payload)
    fmt.Printf("%s\n", metric) # C myap.frobble.count 1 [app:myapp hostname:myhost]
}
```
