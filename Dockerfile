# syntax=docker/dockerfile:1
FROM golang:1.24.1 AS build
WORKDIR /projects
COPY . .
RUN go build -o /associative_entity_redux

FROM gcr.io/distroless/base-debian12
COPY --from=build /associative_entity_redux .
