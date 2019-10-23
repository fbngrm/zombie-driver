FROM golang:1 AS build
COPY . /workspace
WORKDIR /workspace
ARG _TAG
ENV VERSION ${_TAG}
WORKDIR make build

FROM debian:stable-slim
COPY --from=build /workspace/bin/zombie-driver /bin/zombie-driver
EXPOSE 8082 9104
CMD . /secrets/env-test; /bin/zombie-driver

