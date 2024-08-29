# Stream Homework assignment

## Introduction

### Prerequisites

- You'll need Go 1.22 or newer installed. See https://go.dev/doc/install.
- Docker for setting up databases.

### Setup

First, start the docker compose setup. This provides a PostgreSQL and Redis
that we can connect to.

```
docker compose up -d
```

### Running the API server

We can run the API with `go run`. The flags will by default point to the DBs
defined in `docker-compose.yml`.

```
go run ./cmd/api
```

This will start listening on port `8080`:

```
curl -s localhost:8080
```

_See `go run ./cmd/api -h` for flags_

### Running tests

Unit tests can be run directly with `go test`:

```
go test ./...
```

#### Integration tests

> Note: running integration tests will wipe all existing data in the database.

Integration tests that use real databases are opt-in. This is partially so that
tests can be run without Docker, but also because the tests will remove all
existing data from the DB. If the docker-compose stack is up, we can opt into
integration tests with the `-integration` tag:

```
go test -tags=integration ./...
```

The tests will point to the same databases in docker compose.

#### End-To-End tests

A small [Hurl] script allows simulating real requests to the API. 

> The test assumes an empty starting point, so you may need to restart the
> docker compose stack if data was previously inserted.

With Hurl installed and the API running, we can run these tests:

```
hurl --test e2e.hurl
```

## Assignment

This homework assignment is part of the interview process (for backend engineers) at Stream (btw we're [hiring](https://getstream.io/team/#jobs)). It is a simplified version of the real-world problems we solve every day.
For this assignment we will provide you with a simple message (REST) API. The goal of the assigment is for you to add missing functionality.

The way the API is currently implemented, it's possible to create messages and retrieve a list of messages. We have also included an endpoint to add a reaction to a message. However, this endpoint is not implemented yet.

A message consists of the following fields:
* `ID` - a unique identifier for the message
* `Text` - the content of the message
* `UserID` - the user who created the message
* `CreatedAt` - the timestamp when the message was created

The API has three endpoints:
* `GET /messages` - returns a list of messages
* `POST /messages` - creates a new message
* `POST /messages/{messageID}/reactions` - adds a reaction to a message

When a message is created (`POST /messages`), the message is inserted into the database (`messages` table). The message is also added to the message cache (Redis). The message cache is used to store the latest 10 messages. When a message is added to the cache, the oldest message is removed if the cache is full (ie. contains a maximum of 10 messages).
When a list of messages is retrieved (`GET /messages`), the application first tries to fetch the messages from the message cache. If the cache does not contain the requested messages, the application fetches the messages from the database.

Your task is to implement the missing functionality. These are the things we would like you to do:
1. Create the database schema for the `reactions` table (`postgres/schema.sql`).
2. We assume this API will be heavily used and therefor needs to support lots of read and write operations to the database. Update the schema accordingly.
3. Implement the `POST /messages/{messageID}/reactions` endpoint. A reaction consists of the following fields:
    * `ID` - a unique identifier for the reaction
    * `MessageID` - the message to which the reaction is added
    * `UserID` - the user who added the reaction
    * `Type` - the type of the reaction (eg. like, love, laugh, etc.)
    * `Score` - the score of the reaction (eg. 1, 2, 10, etc.) If no score is provided, the default score is 1. You can think of the score as claps on Medium.com.
    * `CreatedAt` - the timestamp when the reaction was added.
4. Extend the list messages endpoint (`GET /messages`) to allow for pagination. The endpoint currently tries to fetch all messages. Long term this is not scalable, therefore we would like to be able to fetch messages in pages of 10 messages. For example, the first page should return the latest 10 messages, the second page should return messages the next 10, etc.
5. Extend the list messages endpoint (`GET /messages`) to include the total reaction count as well as a list of reactions for each message.
6. The endpoints currently don't validate the input. Please add input validation to the endpoints.

We expect a certain level of seniority. Therefor we don't want to explicitly mention every detail you should think about. Think about this assignment as a real-world problem you need to solve. We expect you to make the necessary decisions to make the application (almost) production ready.

### Deliverables
Please fork this repository and implement the missing functionality. An experienced engineer would most likely be able to complete the assignment in a few hours. When you are done, please send us a link to your repository.
We will schedule a follow-up call to discuss your solution.

### Bonus points
The following things are not part of the assignment, but if you feel like making the application production ready, you could do something with the following topics:
* Improve robustness with graceful degradation when databases are down
* Observability

### Reading materials
* [Stream's 10 week onboarding program](https://stream-wiki.notion.site/Stream-Go-10-Week-Backend-Eng-Onboarding-625363c8c3684753b7f2b7d829bcd67a)


[Hurl]: https://hurl.dev/
