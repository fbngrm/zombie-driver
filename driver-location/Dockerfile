FROM golang:1 AS build
COPY . /workspace
WORKDIR /workspace
ARG _TAG
ENV VERSION ${_TAG}
RUN make build

FROM debian:stable-slim
COPY --from=build /workspace/bin/driver-location /bin/driver-location
EXPOSE 8081 9103
CMD . /secrets/env-test; /bin/driver-location
