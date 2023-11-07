FROM golang:1.21 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN make build

FROM golang:1.21

RUN apt-get update && \
    apt-get install -y jq curl unzip

WORKDIR /app

COPY --from=builder /app/bin/leader_elector ./

ENTRYPOINT [ "./leader_elector" ]

