# Task Manager Exercise

This project is just a POC to manage tasks performed during a working day.
There are two roles within this application, Manager and Technician. The manager is able to delete
and see all tasks within the application, while the Technician can view, create and update
it's own tasks. When a task is completed the manager must be notified that a task was done.

This POC needs to be integrated in a micro service ecosystem and it's out of the scope of this POC
user identity management or AUTHN.

## Project Structure

This project tries to follow the [golang project layout](https://github.com/golang-standards/project-layout).

* /api - Api documentation
* /build - Needed files to build the application
* /cmd - Main files, for both server and notification client.
* /deployments - K8s files
* /internal - Specific pkg for the application
* /internal/config - Configuration helper, currently used to collect Env variables
* /internal/repository - All repository storage related files, Gorm, Storage abstraction, Task repo and Task Query
* /internal/server - REST API server and handler files
* /internal/service - Business logic to handle tasks
* /pkg - Packages that can be imported if other applications want to reuse this code
* /pkg/dto - Data Transfer Objects
* /pkg/encryption - AES encryption files
* /pkg/logwrapper - Log configuration
* /pkg/models - Application models
* /pkg/rabbitmq - Broker implementation

## Architecture

This applications uses a REST interface to allow users to CRUD Tasks. The tasks are stored on a DB, this interaction is done
through a task repo. When a task is updated to a complete status a message is sent to a channel, via AMQP/RabbitMQ.

![Arch](docs/arch.drawio.png)

## Requirements

* [golangci](https://github.com/golangci/golangci-lint) - Fast Go linters runner
* [gosec](https://github.com/securego/gosec) - Inspects source code for security problems by scanning the Go AST.
* [swag](https://github.com/swaggo/swag) - Converts Go annotations to Swagger Documentation 2.0.
* [docker-compose](https://docs.docker.com/compose/) - Compose is a tool for defining and running multi-container Docker applications.
* [RabbitMQ](https://github.com/streadway/amqp) - RabbitMQ client to publish and consume messages.
* [zap](https://github.com/uber-go/zap) - Logging library
* [gin](https://github.com/gin-gonic/gin) - API with performance that is up to 40 times faster thanks to httprouter.

## REST API

The REST API is located under internal/server. It leverage [gin web framework](https://github.com/gin-gonic/gin)

### Auth

The API authentication is done from an identity provider, such as keycloak. This service must receive a valid JWT
as part of the authentication header (with the bearer prefix).

This JWT must have the sub claim and the roles claim.

### Middleware

In order to process the requests in a clean way and using a DRY approach this project uses some middleware. One to evaluate JWT other RBAC and finally
one to trace requests.

#### JWT Middleware
There's a middleware that extracts the needed JWT information. The following table highlights the needed claims

| Claim      | Description                                                                            |
|----------|------------------------------------------------------------------------------------------|
|  roles   | A list with the user's roles, e.g, "Roles": ["Manager","Technician"]                     |  
|  sub     | Claim that contains the user id that will be used to evaluate object ownership           |

If other claims are needed first evaluate the [public claims registered by the IANA](https://www.iana.org/assignments/jwt/jwt.xhtml)

#### RBAC Middleware
The RBAC middleware validates if a given user has the needed role to access a given endpoint. The middleware accepts a comma separated roles list
that will be allowed to request the endpoint.

#### Request Track Middleware
There's a middleware that will create a request identifier for each request performed. This will be used on the business logic in order
to allow request tracking and better debugging. This middleware adds to the context a x-request-id key to the context and as a response header.


### JWT Examples

The following JWT headers can be used to test this application


Tech role with user id 74531653-252b-48c7-b562-63e82f5e3466 
```console
Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJPbmxpbmUgSldUIEJ1aWxkZXIiLCJpYXQiOjE2NDU0ODU1MTYsImV4cCI6MTY3NzAyMTUxNiwiYXVkIjoidGFzay1zZXJ2aWNlIiwic3ViIjoiNzQ1MzE2NTMtMjUyYi00OGM3LWI1NjItNjNlODJmNWUzNDY2Iiwicm9sZXMiOiJUZWNobmljaWFuIn0.akelYA5OTpqh0Fvto1J2Rp4FQkzNCQKm68xm_FLzYqY
```

```json
{
  "iss": "Online JWT Builder",
  "iat": 1645485516,
  "exp": 1677021516,
  "aud": "task-service",
  "sub": "74531653-252b-48c7-b562-63e82f5e3466",
  "roles": "Technician"
}
```

Manager role with user id 69c10abf-487f-42b3-8730-a58947298fe5
```console
Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJPbmxpbmUgSldUIEJ1aWxkZXIiLCJpYXQiOjE2NDUyMjg4MDAsImV4cCI6MTY3Njc2NDgwMCwiYXVkIjoidGFzay1tYW5hZ2VyLXBvYyIsInN1YiI6IjY5YzEwYWJmLTQ4N2YtNDJiMy04NzMwLWE1ODk0NzI5OGZlNSIsInJvbGVzIjoiTWFuYWdlciJ9.l8cVqY0dOXj5-jVlGPjl2B5CQ0q4QDgDfO9z_7XOvwo
```

```json
{
  "iss": "Online JWT Builder",
  "iat": 1645228800,
  "exp": 1676764800,
  "aud": "task-manager-poc",
  "sub": "69c10abf-487f-42b3-8730-a58947298fe5",
  "roles": "Manager"
}
```

Manager and Tech roles with user id 69c10abf-487f-42b3-8730-a58947298fe5
```console
Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJPbmxpbmUgSldUIEJ1aWxkZXIiLCJpYXQiOjE2NDUyMjg4MDAsImV4cCI6MTY3Njc2NzU0MSwiYXVkIjoidGFzay1tYW5hZ2VyLXBvYyIsInN1YiI6IjY5YzEwYWJmLTQ4N2YtNDJiMy04NzMwLWE1ODk0NzI5OGZlNSIsInJvbGVzIjpbIk1hbmFnZXIiLCJUZWNobmljaWFuIl19.ZRmU23SX0m3FpRv7ptEc4252pZYRn_dGPC-ayB8OMEs
```
```json
{
  "iss": "Online JWT Builder",
  "iat": 1645228800,
  "exp": 1676767541,
  "aud": "task-manager-poc",
  "sub": "69c10abf-487f-42b3-8730-a58947298fe5",
  "roles": [
    "Manager",
    "Technician"
  ]
}
```

Different tokens can be built using an [online tool](http://jwtbuilder.jamiekurtz.com/)

#### Roles

Currently the application uses two distinct roles: (i) Manager and (ii) Technician.
The Manager is able to view and delete the tasks from all technician.
The Technician can see, create and update the tasks. 

### Endpoints

* [GET] /tasks - Allow to filter specific tasks via query parameters
* [POST] /tasks - Allow to create a new task
* [GET] /tasks/task_id - Allow to collect all information about a specific task
* [PATCH] /tasks/task_id - Allow to set the completed date or change the summary
* [DELETE] /tasks/task_id - Allow to delete a specific task

All the role access of each endpoint is explained on the roles section and all the dto needed to interact with this api are described on
the documentation.

A valid JWT token must be provided. The signature is ignored since it's validation is out scope of this project.

Since the task summary can contain PII information this column must be encrypted.
Because this application is the one doing both encryption and decryption the algorithm used is [AES](https://pt.wikipedia.org/wiki/Advanced_Encryption_Standard). The encryption is done
on both PATCH and POST endpoints.

### Configuration

| Key           | Description                                                      |
|---------------|------------------------------------------------------------------|
|  AES_SECRET   | The secret that will be used to encrypt and decrypt PII data     |  
|  SERVER_PORT  | The port where the API will be running                           |

### Documentation

The API is document with OpenAPI Specification. It uses swagger to allow a better interaction between the client and the API for testing purposes.
The documentation is stored under api/docs and its generated via swag. There's a make option to generate the documentation.

```console
$ make swagger
```

The swagger endpoint by default is located on http://0.0.0.0:8000/docs/index.html#/ it is used only for development purposes.
## Message Broker

The Message broker is a component that allows to manage the messages across multiple services. The broker used on this project is the rabbitmq.

The management interface can be access in http://localhost:15672/

To know more about RabbitMQ internals please read the following [link](https://www.rabbitmq.com/tutorials/amqp-concepts.html) 

### Auth

The auth of the broker relies on the one implemented by rabbitmq. Currently accessing the virtual host grants access to all messages,
to change this behavior some configurations are needed, to know more use the following [link](https://www.rabbitmq.com/access-control.html) 

### Exchange

The exchange is configured with the default values. An improvement is to study which ones might be the best for the project.

### Producer

The producer produces the messages that will be consumed by the consumers. The producer uses the application/json on the messages and sends them
via direct exchange through a routing key. 

### Consumer

The consumer must implement a function that will handle the received messages. When receiving a message the handler will process it and if the processing
is successful the msg will be acknowledge.

### Configuration

| Key        | Description                                                                        |
|------------|------------------------------------------------------------------------------------|
|  RMQ_ADDR  | The host where the RMQ is located                                                  |  
|  RMQ_USER  | The user that will use RMQ                                                         |  
|  RMQ_PWD   | The user's password                                                                |  
|  RMQ_PORT  | The port where RMQ is hosted                                                       |  
|  RMQ_VHOST | The [virtual host](https://www.rabbitmq.com/vhosts.html) used by the application   |
| TASKS_COMPLETE_ROUTING | The routing key used to send notification about completed tasks        |

Both exchange and queue are named as "tasks".

## Database

Currently this project needs a MySQL database to persist data.
To abstract the communication between the application and the database it uses [GORM](https://gorm.io/) [ORM](https://en.wikipedia.org/wiki/Object%E2%80%93relational_mapping).

To abstract the GORM layer it is used a storage interface that must implement the methods to enable CRUD operations and a migrate to apply migrations and a close to close the db connection.

Currently the project has the following tables:
* Task - Stores the task information and the time when it was concluded

### Configuration

The DB configuration is done through Env Variables.

| Key      | Description                                 |
|----------|---------------------------------------------|
|  DB_ADDR | The host where the DB is located            |  
|  DB_USER | The db user, it must have access to the db  |
|  DB_PWD  | The user's password                         |
|  DB_PORT | The port to access the database             |
|  DB_Name | The name of database to use                 |

## Logging

The application uses [zap](https://github.com/uber-go/zap) to perform logging.

It's used the most performant logger.
The application uses two loggers, one that logs directly to stdout and other that stores the logs within a log file.

### Configuration

The logger level is done through Env Variables.

| Key           | Description                                                      |
|---------------|------------------------------------------------------------------|
|  LOG_LEVEL    | One of Info or Debug, If value is invalid Info is used.          |  

## Deployment

The project has a Dockerfile that allows to create a container for both server and client. There's a development environment that can be 
used through the following make command.

```console
$ make dev-env-up
```

To stop the there are two options, (i) that keeps the volumes and other (ii) that clean the volumes

```console
$ make dev-env-down
```

```console
$ make dev-env-down-volumes
```

There are also some kubernetes files to demo a possible deployment. It was not included any infrastructure, such as mysql or rabbitmq as well as secrets files. 

## Tests

The tests within this project are separated by tags.
* Unit - Only unit tests
* Integration - When integrating one or more components

The integration tests run on a docker env:

```console
$ make test-all
```

To run only the unit tests:

```console
$ make test-unit
```

It's possible to collect coverage html report by using the following make command:

```console
$ make coverage-html
```

More work is required on the tests, as the language is not uniform and there are some parts of the code that
could be reused following a DRY approach.
More integration and e2e tests are required to improve the code coverage.

## Githooks

In order to use the commit hooks it's needed the pre-commit command. Then just install the hooks

```console
$ pre-commit install
```

The hooks will try to enforce best practices and avoid security issues.

## Github Pipeline

Currently there is no pipeline configured for this project