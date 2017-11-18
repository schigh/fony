[![Build Status](https://travis-ci.org/schigh/fony.svg?branch=master)](https://travis-ci.org/schigh/fony)
[![Go Report Card](https://goreportcard.com/badge/github.com/schigh/fony)](https://goreportcard.com/report/github.com/schigh/fony)

# fony
A phony endpoint simulator for your integration tests

### How to use
Create a json file with a list of your endpoints like so:

```json
{
  "global_headers": {
    "X-Fony": "true"
  },
  "endpoints": [
    {
      "url": "/foo/bar",
      "method": "GET",
      "responses": [
        {
          "headers": {
            "X-Herp": "DDDERP"
          },
          "payload": {
            "foo": "bar",
            "fizz": "buzz"
          },
          "status_code": 200
        },
        {
          "headers": {
            "X-Herp": "HHHERP"
          },
          "payload": {
            "foo": "bark",
            "fizz": "moo"
          },
          "status_code": 201
        }
      ]
    }
  ]
}
```

For each endpoint, the url and verb are required.  There must also be at least one response per endpoint as well.

If you have multiple responses for the same endpoint, you can specify which one to return in the headers of your request.  To do this, set a header called `X-Fony-Index` to the index of the response you wish to return. If you pass an index out of range, the response at index `0` will be returned.

Use `docker-compose` to run the `fony` command.  Be sure to set the `GOBIN` environment variable to `/go/bin`.
See the `docker-compose.yml` file in this repository for an example.

## Things that need done
- All the tests

## Caveats
This project is meant to run in a Docker container.  The necessary go dependencies will be installed when the Dockerfile is built.  If you want to run this program locally, you will need to install the following dependencies via `go get`:
- `github.com/sirupsen/logrus`
- `goji.io`
