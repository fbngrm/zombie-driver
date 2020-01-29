![zombie](zombie-150.png)

## Setup
This section assumes there is a go, docker, make and git installation available on the system.

To check your installation, run:

```bash
go version
docker version
make --version
git version
```

Fetch the repo from GitHub:

```bash
cd $GOPATH/src/github.com/heetch
git clone git@github.com:heetch/FabianG-technical-test.git
cd FabianG-technical-test
git checkout development
```

### Dependency management
For handling dependencies, go modules are used.
This requires to have a go version > 1.11 installed and setting `GO111MODULE=1`.
If the go version is >= 1.13, modules are enabled by default.
There might be steps required to access private repositories.
If you have problems setting up or building the project which are related to modules, please consider reading up the [documentation](https://github.com/golang/go/wiki/Modules).
If this does not solve the issue please open an issue here.

### Usage
Makefiles are provided which should be used to test, build and run the services separately or all at once.
The services and backing services are started in a docker container.
The configuration resides in the [docker-compose](https://github.com/heetch/FabianG-technical-test/blob/development/docker-compose.yaml) file.
The Dockerfiles used to build images are located in the project root.


### Build
Builds will are located in the `/bin` sub-directory of each service. Binaries use the latest git commit hash or tag as a version.

```bash
make all # builds all services
```

### Run
Services are intended to be ran in a docker container.

```bash
make up # builds docker images and runs all services and backing services.
```

### Tests
There are several targets available to run tests.

```bash
make test # runs tests for all services
make test-cover # creates coverage profiles for all services
make test-race # tests services for race conditions
```

### Lint
There is a lint target which runs [golangci-lint](https://github.com/golangci/golangci-lint) in a docker container.

```bash
make lint
```

### Service level
Except for `up` and `lint`, all targets are available on a service level.
Run the make command from the respective service directory or use the `-C` argument.

```bash
make -C <service_name> all # builds <service_name>
```

### Code changes
After making changes to the code, you need to rebuild the image(s):

```bash
docker-compose up --detach --build <service_name>
```

### Configuration
Services can be configured by parameters or environment variables.
For configuring the services via environment variables use the docker-compose file.
Alternatively, provide arguments to the command directly.


### gateway

| Arg              | ENV            | default |                           | Required |
|------------------|----------------|---------|---------------------------|----------|
| --cfg-file       | CFG_FILE       |         | path to config file       | True     |
| --http-addr      | HTTP_ADDR      |         | address of HTTP server    | True     |
| --metrics-addr   | METRICS_ADDR   |         | address of metrics server | True     |
| --service        | SERVICE        | gateway | service name              | False    |
| --shutdown-delay | SHUTDOWN_DELAY | 5000    | shutdown delay in ms      | False    |
| --version        |                |         | show application version  | False    |

### driver-location

| Arg                       | ENV                    | default         |                                | Required |
|---------------------------|------------------------|-----------------|--------------------------------|----------|
| --cfg-file                | CFG_FILE               |                 | path to config file            | True     |
| --http-addr               | HTTP_ADDR              |                 | address of HTTP server         | True     |
| --metrics-addr            | METRICS_ADDR           |                 | address of metrics server      | True     |
| --redis-addr              | REDIS_ADDR             |                 | address of metrics server      | True     |
| --nsqd-tcp-addrs          | NSQD_TCP_ADDRS         |                 | TCP addresses of NSQ deamon    | True     |
| --nsqd-lookupd-http-addrs | NSQ_LOOKUPD_HTTP_ADDRS |                 | HTTP addresses for NSQD lookup | True     |
| --nsqd-topic              | NSQ_TOPIC              |                 | NSQ topic                      | True     |
| --nsqd-chan               | NSQ_CHAN               |                 | NSQ channel                    | True     |
| --nsq-num-publishers      | NSQ_NUM_PUBLISHERS     | 100             | NSQ publishers                 | False    |
| --nsq-max-inflight        | NSQ_MAX_INFLIGHT       | 250             | NSQ max inflight               | False    |
| --service                 | SERVICE                | driver-location | service name                   | False    |
| --shutdown-delay          | SHUTDOWN_DELAY         | 5000            | shutdown delay in ms           | False    |
| --version                 |                        |                 | show application version       | False    |

### zombie-driver

| Arg                   | ENV                 | default       |                                             | Required |
|-----------------------|---------------------|---------------|---------------------------------------------|----------|
| --http-addr           | HTTP_ADDR           |               | address of HTTP server                      | True     |
| --metrics-addr        | METRICS_ADDR        |               | address of metrics server                   | True     |
| --driver-location-url | DRIVER_LOCATION_URL |               | address of driver-location service          | True     |
| --zombie-radius       | ZOMBIE_RADIUS       |               | radius a zombie can move                    | True     |
| --zombie-time         | ZOMBIE_TIME         |               | duration for fetching driver locations in m | True     |
| --service             | SERVICE             | zombie-driver | service name                                | False    |
| --shutdown-delay      | SHUTDOWN_DELAY      | 5000          | shutdown delay in ms                        | False    |
| --version             |                     |               | show application version                    | False    |

### Logging
The current setup uses a human friendly logging format. Service loggers attach the service name and build version to the log output.

### Bugs
Setting logger on NSQ producers and consumers.
The logger used in the project does not implement the required interface to be used in NSQ.
Thus, logs are a bit polluted.

### Example
Run a basic example from the project root:

```bash
# start all services in a docker container
make up

# publish a location via the gateway service
curl --request PATCH -d '{"latitude": 48.864193,"longitude": 20.350498}' 'http://127.0.0.1:8080/drivers/1/locations'

# check locations via the `internal` driver-location service directly; response data may differ
curl --request GET -i 'http://127.0.0.1:8081/drivers/1/locations?minutes=5'
curl: (7) Failed to connect to 127.0.0.1 port 8081: Connection refused # not reachable from host

# zombie check via the `public facing` gateway service; reponse data may differ
curl --request GET -i 'http://127.0.0.1:8080/drivers/1'

HTTP/1.1 200 OK
Content-Length: 23
Content-Type: application/json
Date: Sat, 26 Oct 2019 11:06:56 GMT
Request-Id: bmq2hk790i5q0u9t1pog

{"id":1,"zombie":true}

# publish more data
curl --request PATCH -d '{"latitude": 48.864193,"longitude": 20.450498}' 'http://127.0.0.1:8080/drivers/1/locations'

# zombie check again
curl --request GET -i 'http://127.0.0.1:8080/drivers/1'
HTTP/1.1 200 OK
Content-Length: 24
Content-Type: application/json
Date: Sat, 26 Oct 2019 11:09:00 GMT
Request-Id: bmq2ij790i5ub07vlkk0

{"id":1,"zombie":false}
```
