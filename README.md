# TOP SEARCHES

[![Go](https://img.shields.io/badge/Go-123?logo=go)](https://go.dev/)
[![Gin](https://img.shields.io/badge/Gin-123?logo=gin)](https://gin-gonic.com/en/)
[![Prometheus](https://img.shields.io/badge/Prometheus-621?logo=prometheus)](https://prometheus.io/)
[![Grafana](https://img.shields.io/badge/Grafana-621?logo=grafana)](https://grafana.com/)
[![Docker](https://img.shields.io/badge/Docker-purple?logo=docker)](https://www.docker.com/)
[![Swagger](https://img.shields.io/badge/Swagger-191?logo=swagger)](https://swagger.io/)
[![gRPC](https://img.shields.io/badge/gRPC-white)](https://grpc.io/)
[![REST_API](https://img.shields.io/badge/REST_API-white)](https://en.wikipedia.org/wiki/REST)

---

# About the project

It's my solution to [Wildberries test task](https://ibb.co/PZKYBmq6). The main goal was to create service for managing
top searches made within last 5 minutes, but you can specify any timespan. There was no clear instructions and 
restrictions, so it was all up to me

The service has three-layer architecture **handler → service → storage**. Service can work with gRPC or REST depending
on specified configuration. Kafka was selected as broker, because it's the most popular and powerful broker

The Swagger documentation for REST, benchmark, Kafdrop, Prometheus metrics with Grafana 
visualization are provided. Main parts has Unit-tests

To test the service by yourself an additional producer service was also created. The producer is a simple API with only
one handpoint just to send searches to a broker. You can switch on the auto sender of searches to broker in configuration

---

# Payload

The payload for the broker

```
{
  "search": "phone",
  "user": "198.51.100.14",
  "timestamp": "2026-05-25T10:30:00.123456789Z"
}
```

Search is the search request made by user, used to track the top searches

User is the ip or any other unique info to distinguish users, used to limit requests

Timestamp is the time, when the search was queried, used to filter old searches

---

# Architecture

## Storage

Because the data lives only limited amount of time there is no database used, it would be not only redundant but also slower due to 
additional requests between server and database. Because service should survive during highload and have low latency
the data is being stored directly in Go via maps and slices

The storage contains main bucket, slice of buckets and current top as slice of strings. Each bucket is a map that stores 
information about searches and number of times when they were queried within specific period, that can be specified in 
configuration. Amount of buckets in slice is being calculated as (timespan / bucket's timespan) + 1, so each timespan is being distributed evenly between buckets. There 
is an additional bucket, that is being used during refreshing the top

Because the service can have planty amount of requests it would be irrational to update the top for every request, so 
that's why the top is being refreshed once in a lifetime of a bucket in a background goroutine. Every period of bucket's 
timespan one of the buckets in slice is being cleared, so the number of active buckets is always one less than total 
number, so using an additional buckets make it possible for evenly distribution of timespan

The top is being arranged by sorting searches in main bucket by their number of appearances. Because calculating total
number of searches by ranging in slice of buckets and summing up their data will be not efficient, the total number 
within the timespan is being stored straightaway in main bucket. When adding new search to the bucket in slice, the 
search is also being added to the main bucket. When clearing the bucket in slice, the according number of appearances
is being removed in main bucket. Because the number of requests for getting top is from 10 to 50 times more than 
requests for adding new search, this schema is more efficient

## Stoplist

Satisfying desire of the marketing department, the stoplist with words banned from top searches can be set up via 
config or env, it can be also dynamically changed via provided API. It would be easier if the stoplist was static 
because in this case we can just block add search request in service layer so it won't be added to storage. But to
make stoplist dynamically we should add search to the storage anyway, because any word can be added to or removed from 
stoplist, so we need to keep track of all searches, and filter them in get request

## Limiter

The desire of analytic department was also satisfied by implementing a limiter. Because we can't separate outlier in 
data due to natural hype and cheating, the limiter filters searches by users and not by the search itself. It is quite 
natural because a real user can't make searches too frequently. So the limiter limits one search request from user within
cooldown specified in configuration

The limiter has the similar realisation as the storage, but there are only two buckets. At every request an user is being
placed in bucket and can't make another until he is being already placed in current bucket. The bucket also clears every 
period of cooldown

## Metrics

Basic metrics (active connections, total requests, request duration, total searches, total blocked searches, total 
stoplist hits and current top size) is being provided via Prometheus. To make the metrics more convenient to use
the Grafana visualization is also being provided

---

# Benchmark

To test the perfomance wrk was used, add search requests was sent via post.lua

## Mild test

You can try it out by yourself using the commands below in two separate terminals at the same time

```bash
wrk -t4 -c100 -d60s -s post.lua --latency http://localhost:8081/search
wrk -t8 -c200 -d60s --latency http://localhost:8080/top-searches/get/5
```

### Writing

Running 1m test @ http://localhost:8081/search

4 threads and 100 connections

| Thread Stats| Avg    | Stdev  | Max      | +/- Stdev |
|-------------|--------|--------|----------|-----------|
| Latency     | 6.70ms | 3.86ms | 132.20ms | 85.16%    |
| Req/Sec     | 3.78k  | 480.10 | 7.81k    | 74.11%    |

Latency Distribution

     50%    6.07ms
     75%    8.20ms
     90%   10.62ms
     99%   16.40ms

902992 requests in 1.00m, 300.54MB read

    Requests/sec:    15044.73
    Transfer/sec:      5.01MB

### Reading

Running 1m test @ http://localhost:8080/top-searches/get/5

8 threads and 200 connections

| Thread Stats| Avg    | Stdev   | Max      | +/- Stdev  |
|-------------|--------|---------|----------|------------|
| Latency     | 4.40ms | 4.01ms  | 132.65ms | 84.85%     |
| Req/Sec     | 6.38k  | 0.93k   | 22.32k   | 73.73%     |

Latency Distribution

     50%    3.59ms
     75%    5.93ms
     90%    9.00ms
     99%   17.09ms

3047895 requests in 1.00m, 1.09GB read
Socket errors: connect 0, read 0, write 0, timeout 25

    Requests/sec:  50715.64
    Transfer/sec:   18.52MB

## Highload test

You can try it out by yourself using the commands below in two separate terminals at the same time

```bash
wrk -t4 -c200 -d60s -s post.lua --latency http://localhost:8081/search
wrk -t8 -c10000 -d60s --latency http://localhost:8080/top-searches/get/5
```

### Writing

Running 1m test @ http://localhost:8081/search
4 threads and 200 connections

| Thread Stats| Avg      | Stdev   | Max     | +/- Stdev  |
|-------------|----------|---------|---------|------------|
| Latency     | 13.15ms  | 6.39ms  | 70.60ms | 73.38%     |
| Req/Sec     | 3.82k    | 835.23  | 12.40k  | 78.37%     |

Latency Distribution

     50%   11.99ms
     75%   16.35ms
     90%   21.34ms
     99%   33.72ms

913385 requests in 1.00m, 304.00MB read

    Requests/sec:    15201.85
    Transfer/sec:      5.06MB


### Reading

Running 1m test @ http://localhost:8080/top-searches/get/5

8 threads and 10000 connections

| Thread Stats| Avg      | Stdev      | Max     | +/- Stdev  |
|-------------|----------|------------|---------|------------|
| Latency     | 267.69ms | 249.20ms   | 2.00s   | 61.63%     |
| Req/Sec     | 5.35k    | 1.23k      | 14.45k  | 72.09%     |

Latency Distribution

     50%  230.43ms
     75%  409.33ms
     90%  583.65ms
     99%     1.13s 

2548441 requests in 1.00m, 0.91GB read
Socket errors: connect 0, read 0, write 0, timeout 879

    Requests/sec:  42410.17
    Transfer/sec:   15.49MB

---

# Downloading and running the api

## 1. Run Docker

Install and run Docker on your computer

## 2. Install the api

Clone the repository

```bash
git clone https://github.com/middelmatigheid/top-searches.git
cd top-searches
```

## 3. Specify configuration

Specify configuration in config.yaml files or environment. Take notes, that the producer has separate config.yaml

Environment configuration

```
PORT = YOUR_PORT
PORT_GRPC = YOUR_PORT_GRPC
GRPC = YOUR_FLAG                                #   set it to "true" if you want to run the service in gRPC mode
TIMESPAN = YOUR_TIMESPAN                        #   top's timespan in seconds
BUCKET_TIMESPAN = YOUR_BUCKET_TIMESPAN          #   bucket's timespan in seconds, make sure to choose a divisor of TIMESPAN
COOLDOWN = YOUR_COOLDOWN                        #   limiter cooldown
BROKERS = YOUR_BROKERS                          #   URLS for brokers, separated by comma
TOPIC = YOUR_TOPIC                              #   broker's topic
STOPLIST = YOUR_STOPLIST                        #   banned words, separated by comma
LOG_FILE = YOUR_LOG_FILE                        #   file for writing logs

PRODUCER_PORT = YOUR_PRODUCER_PORT
PRODUCER_PORT_GRPC = YOUR_PRODUCER_PORT_GRPC
PRODUCER_GRPC = YOUR_FLAG                       #   set it to "true" if you want to run the producer in gRPC mode
PRODUCER_BROKERS = YOUR_PRODUCER_BROKERS        #   URLS for brokers, separated by comma
PRODUCER_TOPIC = YOUR_PRODUCER_TOPIC            #   broker's topic
PRODUCER_AUTO_SENDER = YOUR_FLAG                #   set it to "true" if you want to run the auto sender
PRODUCER_SEARCHES = YOUR_PRODUCER_SEARCHES      #   searches for auto sender, separated by comma
PRODUCER_USERS = YOUR_PRODUCER_USERS            #   users for auto sender, separated by comma
PRODUCER_LOG_FILE = YOUR_PRODUCER_LOG_FILE      #   file for writing logs
```

## 4. Run the containers

### Run the containers

```bash
docker compose up -d
```

To run containers with producer, swagger or kafdrop use profiles

    --profile producer     →   Runs the producer
    --profile swagger-ui   →   Runs the Swagger
    --profile kafdrop      →   Runs Kafdrop

```bash
docker compose --profile producer --profile swagger-ui --profile kafdrop up -d
```

### Stop the containers

Stops all docker containers

```bash
docker stop $(docker ps -q)
```

### Deleting the containers

Be careful using the commands below because it will delete all docker containers and volumes

```bash
docker system prune -a --volumes
docker volume prune -a -f
```

### 5. The service is running

---

# Usage

## gRPC mode

Use this command via terminal

### See running services

```bash
grpcurl -plaintext localhost:9091 list
```

If you are running the producer

```bash
grpcurl -plaintext localhost:9094 list
```

### Get top N

Specify the N instead of 5

```bash
grpcurl -plaintext -d '{"n": 5}' localhost:9091 searches.Searches/GetTopN
```

### Get stoplist

```bash
grpcurl -plaintext localhost:9091 stoplist.Stoplist/GetStoplist
```

### Add word to stoplist

Specify the word instead of "pants"

```bash
grpcurl -plaintext -d '{"word": "pants"}' localhost:9091 stoplist.Stoplist/Add
```

### Remove word from stoplist

Specify the word instead of "pants"

```bash
grpcurl -plaintext -d '{"word": "pants"}' localhost:9091 stoplist.Stoplist/Remove
```

### Search request

If you are running the producer

Specify the search and user instead of "pants" and "198.51.100.14"

```bash
grpcurl -plaintext -d '{"search": "pants", "user": "198.51.100.14"}' localhost:9094 producer.Producer/SendSearch
```

## REST mode

Swagger would be available at http://localhost:8082


## Metrics

### Grafana

Grafana would be available at http://localhost:3000

To add Prometheus to Grafana use http://prometheus:9090

### Kafdrop

Kafdrop would be available at http://localhost:9000

## Run tests

To run Unit-tests use

```bash
go test -v ./...
```

## Proto made with

---

# Project structure

```bash
top-searches/
├── cmd/server/main.go          # Main server to run
├── docs/                       # Swagger docs
├── proto/                      # Proto services for gRPC
├── internal/                   
│   ├── config/                 # Getting config from yaml and env
│   ├── consumer/               # Kafka consumer
│   ├── server/                 # gRPC server
│   ├── handler/                # Rest handler
│   ├── service/                # Service layer with business logic
│   ├── storage/                # Storage for managing the top
│   ├── stoplist/               # Managing the stoplist
│   ├── limiter/                # Searches limiter for blocking spam
│   ├── models/                 # Structs
│   └── prometheus/             # Prometheus metrics
├── post.lua                    # Request template for benchmark
├── config.yaml                 # Server's config
├── prometheus.yaml             # Prometheus' config
├── dockerfile    
├── dockerignore
├── docker-compose.yml         
├── go.mod
├── go.sum
├── README.md
│   
└── producer/                   # PRODUCER
    ├── cmd/server/main.go      # Main server to run
    ├── docs/                   # Swagger docs
    ├── proto/                  # Proto services for gRPC
    ├── internal/                 
    │   ├── config/             # Getting config from yaml and env  
    │   ├── server/             # gRPC server
    │   ├── handler/            # Rest handler
    │   ├── service/            # Service layer with business logic
    │   └── models/             # Structs
    ├── config.yaml             # Producer's config
    ├── dockerfile
    ├── dockerignore   
    ├── go.mod
    └── go.sum  
```
