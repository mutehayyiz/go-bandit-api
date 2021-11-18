# How to use

## Prerequisites

- Redis

## Install

Clone the repository

```
$ git clone https://github.com/mutehayyiz/go-bandit-api
```

### Compile from source

Create and configure `config.conf` file like [config.conf.example](https://github.com/mutehayyiz/go-bandit-api/sample_config.json)

```
$ go build
```

### Docker Compose

Configure `config.conf.docker` file

```
$ docker-compose up -d
```

## Requests 

#### Start Scan 

###### Sample Request 

```
POST /scan HTTP/1.1
Host: localhost:4242
Content-Type: application/json

{
   "url": "https://github.com/example/example",
}

```
###### Sample Response 

```
{
    "id": "5e6ec9ce571983fd0e468213",
}
```

#### Get Result

###### Sample Request 

```
GET /scan/{id} HTTP/1.1
Host: localhost:4242
Content-Type: application/json

```
##### Sample Response 

```
{
    "id": "a863ef5f-7517-4fa0-b94b-e8a4e88c60da",
    "created_at": "2021-11-18T06:31:37.606574837Z",
    "updated_at": "2021-11-18T06:31:37.734063134Z",
    "url": "https://github.com/example/example",
    "status": "*",
    "is_secure": false,
    "result": "*"
    "error": "*"
}
```
