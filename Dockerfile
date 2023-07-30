# syntax=docker/dockerfile:1

FROM golang:1.20-bookworm
WORKDIR /go/src/github.com/dofusdude/doduda/
COPY . .
RUN CGO_ENABLED=0 go build -a -installsuffix cgo -o doduda .

FROM python:3.11-bookworm
WORKDIR /app/
COPY --from=0 /go/src/github.com/dofusdude/doduda/doduda ./
COPY persistent persistent

COPY PyDofus PyDofus

ENTRYPOINT [ "./doduda" ]