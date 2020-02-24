# Overview

```sh
# generate go templates
$ go generate
```

Build and run locally using `docker-compose`:

```sh
# build
$ docker-compose build --force-rm --no-cache --pull

# run
$ docker-compose up -d

# stop
$ docker-compose down
```

## Persistent Data

Create persistent Docker volumes for storage:

```sh
$ docker volume create \
  --opt type=none \
  --opt device=/path/to/search-postgresql \
  --opt o=bind search-postgresql
```
