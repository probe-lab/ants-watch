# Ants Watch

Celestia Lightnode Population Monitor.

<img src="./resources/ants.png" alt="Ants Watch" height="300"/>

Authors: [guillaumemichel](https://github.com/guillaumemichel), [kasteph](https://github.com/kasteph)

## Setup

Before installing dependencies:

``` shell
git submodule init
git submodule update --init --recursive --remote
```

Then, `go mod tidy`.

You'll also need to install some tools: `make tools`.

Migrations can now be applied: `make migrate-up`.
