![zombie](zombie-150.png)

This branch implements the requirements defined in the [task](https://github.com/heetch/FabianG-technical-test/blob/development/REQUIREMENTS.md) description.
Additional functionality is:

* configuration via env variables or command-line args
* instrumentation
* circuit-breaker
* docker containers
* configurable zombie-driver identification `Predicate/business rules`

## Setup
This section assumes there is a go, docker, make and git installation available on the system.

To check your installation, run: 

    go version
    docker version
    make --version
    git version


Fetch the repo from GitHub:


    git clone git@github.com:heetch/FabianG-technical-test.git


### Dependency management
For handling dependencies, go modules are used.
This requires to have a go version > 1.11 installed and setting `GO111MODULE=1`.
If the go version is >= 1.13, modules are enabled by default.
There might be steps required to access private repositories.
If you have problems setting up or building the project which are related to modules, please consider reading up the [documentation](https://github.com/golang/go/wiki/Modules).
If this does not solve the issue please open an issue here.

## Usage
Makefiles are provided which should be used to test, build and run the services separately or all at once.
The services and backing services are started in a docker container.
The configuration resides in the [docker-compose file](https://github.com/heetch/FabianG-technical-test/blob/development/docker-compose.yaml).
The Dockerfiles used to build images are located in the project root.


### Using make
The services and backing services are started in a docker container.

#### Build
Builds will be located in the `/bin` sub-directory of each service.


    make all # builds all services


#### Run
Services are intended to be ran in a docker container.


    make up # builds docker containers and runs all services and backing services


#### Tests
There are several targets available to run tests.


    make test # runs tests for all services
    make test-cover # creates coverage profiles for all services
    make test-race # tests services for race conditions


#### Lint
There is a lint target which runs golangci-lint in a docker container.


    make lint


#### Run on a service level
Except for `up`, all targets are available on a service level.
Run the make command from the respective service directory or use the `-C` argument.


    make -C <service_name> all # builds <service_name>


## Services
Services can be configured by parameters or environment variables.
For configuring the services via env variables use the `docker-compose.yaml`.
Alternatively, provide arguments to the command directly.


### gateway

| Arg              | ENV            | default |                           | Required |
|------------------|----------------|---------|---------------------------|----------|
| --cfg-file       | CFG_FILE       |         | path to config file       | True     |
| --http-addr      | HTTP_ADDR      |         | address of HTTP server    | True     |
| --metrics-addr   | METRICS_ADDR   |         | address of metrics server | True     |
| --service        | SERVICE        | gateway | service name              | False    |
| --shutdown-delay | SHUTDOWN_DELAY | 5000    | shutdown delay in ms      | False    |
| --version        |                |         | Show application version  | False    |

### Bugs
Setting logger on nsq producers and consumers.
The logger used in the project does not implement the required interface to be used in nsq.
Thus, logs are a bit polluted.

## Architecture approach
I mostly followed the go [conventions](https://golang.org/doc/code.html) and [proverbs](https://go-proverbs.github.io/) as well as the [12 Factor-App](https://12factor.net/) principles.

The interfaces are kept small to bigger the abstraction.
Variable names are short when they are used close to their declaration.
They are more meaningful if they are used outside the scope they were defined.
Errors are used as values.

Furthermore, I followed the dependency injection and fail early approach, with very few exceptions.
Components are provided all dependencies they need during instanciation.
The result is either a functioning instance or an error.
On application start-up and error results in termination.
Runtime errors do not lead to a crash or panic.

Configuration is injected at start-up.

Termination signals lead to a graceful shutdown.
Meaning, all servers and handlers stop accepting new requests, process their current workload and shutdown.
Though, there is a configurable shutdown timeout, which may prevent this.

### Instrumentation
Only response time metrics and redis method calls are collected as an example of instrumentation.
In a real world application the run time behavior would be monitored in a more detailed way.
Prometheus is used for aggregating the metrics, which are provided by an http handler to be scraped by an prometheus collector.
A graceful shutdown of the server ensures that the metrics will be scraped eventually by checking against an access counter.
Though, there is a configurable shutdown timeout which prevents this from being guaranteed.

### Tests
I wrote unit tests for core functionality, things expected to break and for edge/error cases.
In general I think testing on package boundaries as well as core functionality internally is a better approach than just aiming for a certain percentage of coverage.
Regarding a few error cases, test coverage should be increased though.

Tests that require and nsq server use a helper script to start and shutdown a docker instance in the background but stream logs to a file to not obfuscate test log.

#### Testdata
There is a testdata directory which provides slightly realistic sample data used in most tests.

#### Redis
Since there are two redis commands used only, I implemented an interface to provide a simple mock in tests.
This requires an extra layer of abstraction which could be avoided using a redis mock library.
If there will be more commands used, I would prefer to add a dependency and remove the abstraction.

### Dependencies
In general, the code is written in a way to use as few dependencies as possible and make use of the standard library whenever possible.
This applies also for tests, where no external libraries are used since they mostly do not provide significant advantages but may obfuscate clear readability.
This especially applies to BDD (Behavioral Driven Design/Development) test libraries, which often introduce test-induced design damage by needless indirection and conceptual overhead.

#### Shared libraries
There are a few shared libraries at the project root which are used in all three services.
I would tend to move each service to an own repo and copy over the library code along.
Although, using go modules with versioning, providing the ability to update incrementally, makes it easier to handle shared libraries.

### Circuit-breaker
The circuit-breaker should additionally be implemented as middleware for HTTP handlers.
This requires a refactoring of the middleware which is out-of-scope during this test.
Note, that is no circuit-breaker applied in the [gateway HTTP proxy handler](https://github.com/heetch/FabianG-technical-test/blob/development/gateway/server/handler.go#L68-L78).

### (Zombie) Workflow
Since I am the only contributer (except for initial commits) and there will be a single PR, I followed a rather pragmatic git workflow.
I mostly implemented features in separate branches though. In the beginning.

## Todo
* fix nsq logging
* fix driver-location url in zombie-driver config
* prometheus scraper + dashboard
* https (gateway, nsq)
* authentication/sessions
* load testing
* tracing
* rewrite middleware
* custom proxy/circuit-breaker
* benchmarks
* resilience
* load-balancer

## License
-

---

> Everything should be made as simple as possible, but not simpler. - A. Einstein

