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
      "verb": "GET",
      "response_payload": {
        "foo": "bar",
        "fizz": "buzz"
      },
      "response_code": 200
    },
    {
      "url": "/herp/derp",
      "verb": "POST",
      "response_payload": [
        {
          "duck": 1,
          "cow": true,
          "pig": null
        }
      ],
      "response_code": 200
    },
    {
      "url": "/i/am/a/teapot",
      "verb": "GET",
      "response_payload": {},
      "response_code": 418
    }
  ]
}
```

Use `docker-compose` to run the `fony` command.  Be sure to set the `GOBIN` environment variable to `/go/bin`.
See the `docker-compose.yml` file in this repository for an example.

## Things that need done
- All the tests

## Known bugs
Headers are not parsing properly at the moment

## Future enhancements
Right now there can only be one call per endpoint/http verb combo.
We need to be able to allow multiple responses for each endpoint/http verb.

One proposed solution would be push a list of response payloads to the endpoint, and return them sequentially.
I will revisit this if the need arises.
