FROM alpine:3.20 AS build

RUN apk add --no-cache go gcc musl-dev git

WORKDIR /mnt
ENV PATH="/root/go/bin/:$PATH"
COPY mise.toml /mnt/
RUN wget -O- https://mise.run | MISE_INSTALL_PATH=/usr/local/bin/mise sh \
	&& mise settings experimental=true \
	&& mise trust \
	# https://github.com/mattn/go-sqlite3/issues/1164#issuecomment-1635253695
	&& CGO_CFLAGS="-D_LARGEFILE64_SOURCE" mise install

# take advantage of cache for go dependencies
COPY go.mod go.sum ./
RUN go mod download

# now copy sources and build, check .dockerignore for files not included
COPY . .
RUN mise exec -- mrake build:dist

FROM alpine:3.20

COPY --from=build /mnt/d2s /usr/local/bin/d2s
COPY d2s.example.toml /etc/d2s/base.toml
EXPOSE 8080
USER nobody:nobody
ENTRYPOINT ["d2s"]
