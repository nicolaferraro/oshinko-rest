# oshinko-rest
REST api for a spark cluster management app

## Development notes

This project is using an OpenAPI definition for its API, located at
`api/swagger.yaml`. The code in this project has been mostly generated by
using the tooling from (go-swagger)[https://github.com/go-swagger/go-swagger].

Due to the generated nature of this codebase, there are a few places to
investigate when looking to add functionality:

* `restapi/configure_oshinko_rest.go`, this file is generated by the go-swagger
  tooling, but is deemed safe to edit as the entry point for endpoint handlers.

* `handlers/*`, this package has been added to help separate the endpoint
  handler functions. New handlers, or handler functionality, should be added
  here.

### Requirements for building

* godep, https://github.com/tools/godep
* go-swagger, https://github.com/go-swagger/go-swagger (only needed for
  validating the api file)
