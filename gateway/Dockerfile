FROM golang:1 AS build
COPY . /workspace
WORKDIR /workspace
ARG _TAG
ENV VERSION ${_TAG}
WORKDIR make build

FROM debian:stable-slim
COPY --from=build /workspace/bin/gateway /bin/gateway
COPY --from=build /workspace/secrets /secrets
EXPOSE 8080 9102
CMD . /secrets/.env-test; /bin/gateway
