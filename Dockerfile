FROM alpine:3.20 AS build

RUN apk add --no-cache go
WORKDIR /mnt
ENV PATH="/root/go/bin/:$PATH"
RUN go install github.com/a-h/templ/cmd/templ@v0.2.778

# take advantage of cache for go dependencies
COPY go.mod go.sum /mnt/
RUN go mod download

# now copy sources and build, check .dockerignore for files not included
COPY . .
RUN TEMPL_EXPERIMENT=rawgo templ generate
RUN go build -ldflags '-s -w -X main.DefaultConfigPath=/etc/d2s/base.toml' -o d2s main.go

FROM alpine:3.20

COPY --from=build /mnt/d2s /usr/local/bin/d2s
COPY d2s.example.toml /etc/d2s/base.toml
EXPOSE 8080
USER nobody:nobody
ENTRYPOINT ["d2s"]
