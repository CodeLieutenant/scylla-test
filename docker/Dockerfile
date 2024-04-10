FROM golang:1.22 as base

WORKDIR /app

COPY . .

ARG DEBUG=on
ENV DEBUG=${DEBUG}

RUN apt-get update \
    && apt-get upgrade -y \
    && apt-get install gcc g++ make -y \
    && GOBIN=/bin go install github.com/go-task/task/v3/cmd/task@latest \
    && GOBIN=/bin task cli-tools

FROM base as debug

RUN task build DEBUG=${DEBUG}

EXPOSE 2345

ENTRYPOINT [ "dlv",  "exec", "--log", "--log-output=debugger,debuglineerr,gdbwire,lldbout,rpc", "--listen=0.0.0.0:2345", "--allow-non-terminal-interactive", "--headless=true", "--api-version=2", "--accept-multiclient", "/app/bin/scylladbtest", "--" ]

FROM base as air

ENTRYPOINT [ "air", "-c", "air.toml" ]