# airspace 

A Concourse extension CLI that wraps fly-builds.

## Requirements

* Go 1.6+
* ANSI terminal

## Usage

```bash
$ go get github.com/patrickrand/airspace

$ airspace --help
Usage of ./airspace:
  -c int
    	count (default 10)
  -p string
    	pipeline/job regex
  -t string
    	target (default "local")

$ airspace -t dev -c 10
...
```