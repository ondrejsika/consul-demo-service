[Ondrej Sika (sika.io)](https://sika.io) | <ondrej@sika.io>

# Consul Demo Service

    2020 Ondrej Sika <ondrej@ondrejsika.com>
    https://github.com/ondrejsika/consul-demo-service

Example service for Consul service discovery

## Example Usage

### Config Environment Variables

- `CONSUL_HTTP_ADDR` - default: `http://127.0.0.1:8500`, example: `http://consul-agent-0:8500`
- `REGION` - default: `default`, example: `us-west`
- `INSTANCE` - default: `0`, example: `1`
- `PORT` - default: `80`, example: `81`

### HTTP Endpoints

- `/` - return message from config stored in Consul KV
- `/refresh` - refresh config from Consul KV
- `/livez` - healthcheck endpoint

### Docker

#### Build Docker Image

```
make build
```

#### Push to Docker Hub

```
make push
```

#### Build & Push

```
make
```

#### Run

```
make up
```

See <http://127.0.0.1>

#### See Logs

```
make logs
```

#### Stop

```
make down
```
