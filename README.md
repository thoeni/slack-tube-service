Build: [![CircleCI](https://circleci.com/gh/thoeni/slack-tube-service.svg?style=svg)](https://circleci.com/gh/thoeni/slack-tube-service)

# Build & Run

## Run locally:
### Pre-requisite:
 - Golang installed (<https://golang.org/>)

### Run:
```go run *.go```

In alternative:
```go build && ./slack-tube-service```

### Access:
Just connect to ```localhost:1123```

## Run within Docker container
### Pre-requisite:
 - Docker daemon running on the local machine or on a docker host (<https://www.docker.com/products/docker-toolbox>)

### Build:
```docker build -t slack-tube-service .```

### Run:
```docker run -it --rm --name slack-tube-service -p 1123:1123 slack-tube-service```

### Access:

#### If docker daemon is running on localhost
Just connect to ```localhost:1123```

#### If using docker machine
Find out your IP: ``` docker-machine ip `docker-machine active` ```
Connect to that IP, port ```1123```.

# APIs
 - ```GET /api/tubestatus/``` -> retrieve status for all the lines
 - ```GET /api/tubestatus/{line}``` -> retrieve status for a specific line (e.g. ```GET /api/tubestatus/Bakerloo```)
 - ```POST /api/slack/tubestatus``` -> retrieve status for all the lines with slack-friendly formatting (uses auth token to validate slack client)
 - ```PUT /api/slack/token/{token}``` -> adds a *slack token* to the authorised list
 - ```DELETE /api/slack/token/{token}``` -> removes a *slack token* from the authorised list

### This is what the integration looks like:
![Slack Integration](http://www.antoniotroina.com/downloads/tube.png)

#What if I just want to use it?
This project is still experimental but it works, and if you want to try this integration for free just drop me an email and I'll get in touch.
