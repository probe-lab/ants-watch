# Ants

Celestia Lightnode Population Monitor.

Author: [guillaumemichel](https://github.com/guillaumemichel)

## Setup

Before installing dependencies:

``` shell
$ git submodule init
$ git submodule update
```

Then, `go mod tidy`.

You'll also need to install some tools: `make tools`.

Migrations can now be applied: `make migrations-up`.
