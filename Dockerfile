#ARG GO_VERSION=1.14.3
FROM golang:1.16-alpine as Audiophile

# Set necessary environmet variables needed for our image

#ARG databaseName
#ARG host
#ARG port
#ARG user
#ARG password
#
#ENV databaseName=$databaseName
#ENV host=$host
#ENV port =$port
#ENV  user = $user
#ENV password=$password


ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /server

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o bin/Audiophile cmd/main.go

FROM scratch

WORKDIR /

COPY --from=Audiophile /server/bin .
COPY --from=Audiophile /server/database/migrations ./database/migrations

EXPOSE 8080
ENTRYPOINT ["./Audiophile"]