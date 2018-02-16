[![Build Status](https://travis-ci.org/schigh/fony.svg?branch=master)](https://travis-ci.org/schigh/fony)
[![Go Report Card](https://goreportcard.com/badge/github.com/schigh/fony)](https://goreportcard.com/report/github.com/schigh/fony)

# fony
fony is an endpoint simulator for use in integration tests.  In a microservice architecture, it is often necessary to spin up several dependent services in order to test your service.  This can quickly become a service dependency nightmare, as you'll then need to bring up the dependencies of your dependencies, and so on, until you end up having to spin up your entire environment.  Being forced to do this will probably make you sad and frustrated.

Great news!  With fony, you only need to spin up a super-ultra-mega lightweight container to fully mock a service.  All you have to do is create one config file per service.  You don't even need to put your config in your project (unless you want to, in which case 👍)...you can just host it on a CDN or something.  Can you believe it?

## usage
fony pairs best with an orchestration tool such as `docker-compose` or `kubernetes`, but it's just as easy to use it by itself.  Using the sample config, run the container:
<pre>
docker run -p 8080:80 \
-e "SUITE_URL=https://raw.githubusercontent.com/schigh/fony/master/sample/sample.json" \
schigh/fony:latest
</pre>

Then visit the super useful endpoint at `http://localhost:8080/foo/bar`

You can also add it directly to your orchestration config.  For example in `docker-compose.yml`:

```yml
myfakeservice:
	image: schigh/fony
	environment:
		- SUITE_URL=https://raw.githubusercontent.com/schigh/fony/master/sample/sample.json
```

You could also include a local suite file from within your project:

```yml
myfakeservice:
	image: schigh/fony
	volumes:
		- /path/to/suite/file.json:/go/src/github.com/schigh/fony/fony.json
```

The fony container looks for a file `fony.json` in its `WORKDIR`. 

## The suite file
The suite file contains all the configuration you'll need to properly mock a service in fony.  Your fony suite is a json object that contains these top-level keys:

- `endpoints` (_list_, required): These are the endpoints to the service.  At least one endpoint must be defined.  Endpoints are defined below.
- `global_headers` (_object_, optional): This is a simple key-value object that defines all the headers returned to your service in fony responses.  They are returned for every request, unless they are overridden in an individual response definition.  If you define a global header here, and then you define the same header in an individual response, that new value will be used for the header.  If you set it to an empty string in an individual response, then the header will be removed from the response.

The endpoint object has the following fields:

- `url` (_string_, required): this is the url that defines this endpoint.  fony uses the very excellent [Goji](https://goji.io) library as a muxxer, so you can use patterns if you wish.  Note that fony doesnt actually do any processing on parsed url data, but the flexibility of the Goji muxxer allows you to pass in requests that match a pattern versus a static url.
- `method`(_string_, defaults to `GET`): this is the HTTP method used in the request to the endpoint.
- `responses` (_list_, required) This is a list of the responses for a given endpoint.  An endpoint may haveone to many responses.  Responses are defined below.
-  `run_sequential` (_bool_): If this flag is set true, then responses will be returned sequentially if there is more than one response for the endpoint.

The response object has the following fields:

- `headers` (_object_): This is a key-value map of individual headers for the response.  If you include a header here that is also defined as a global header, this value will override the global.  If you override a global header with an empty string, the global header will be removed from the response.
- `payload` (mixed): this is the payload that is returned in the response



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

## Using a remote suite file
Fony can use a remote suite file, provided it's accessible via GET at runtime.  To use this option, use the `SUITE_URL` environment variable, e.g.
```
> SUITE_URL=https://raw.githubusercontent.com/schigh/fony/master/sample/sample.json go run fony.go
```
or
```
> docker run -e "SUITE_URL=https://raw.githubusercontent.com/schigh/fony/master/sample/sample.json" schigh/fony
```

## Caveats
This project is meant to run in a Docker container.  The necessary go dependencies will be installed when the Dockerfile is built.  If you want to run this program locally, you will need to install the following dependencies via `go get`:
- `github.com/sirupsen/logrus`
- `goji.io`

## Things that need done
- All the tests
