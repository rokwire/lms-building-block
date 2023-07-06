# LMS Building Block

The LMS Building Block provides access to integrated Learning Management Systems (LMS) for the Rokwire platform.

## Documentation
The functionality provided by this application is documented in the [Wiki](https://github.com/rokwire/lms-building-block/wiki).

The API documentation is available here: https://api.rokwire.illinois.edu/lms/doc/ui/index.html

## Set Up

### Prerequisites

MongoDB v4.2.2+

Go v1.16+

### Environment variables
The following Environment variables are supported. The service will not start unless those marked as Required are supplied.

Name|Format|Required|Description
---|---|---|---
LMS_PORT | < int > | yes | Port to be used by this application
LMS_INTERNAL_API_KEY | < string > | yes | Internal API key for invocation by other BBs
LMS_MONGO_AUTH | <mongodb://USER:PASSWORD@HOST:PORT/DATABASE NAME> | yes | MongoDB authentication string. The user must have read/write privileges.
LMS_MONGO_DATABASE | < string > | yes | MongoDB database name
LMS_MONGO_TIMEOUT | < int > | no | MongoDB timeout in milliseconds. Defaults to 500.
LMS_DEFAULT_CACHE_EXPIRATION_SECONDS | < int > | false | Default cache expiration time in seconds. Defaults to 120
LMS_CANVAS_BASE_URL | < url > | yes | Canvas base URL for API calls
LMS_CANVAS_TOKEN_TYPE | < string > | yes | Canvas token type (e.g Bearer)
LMS_CANVAS_TOKEN | < string > | yes | Canvas token that will be used for auth with Canvas APIs
LMS_TEST_USER_ID | < string > | yes | Account ID of test user
LMS_TEST_NET_ID | < string > | yes | Net ID of test user
LMS_TEST_USER_ID2 | < string > | yes | Account ID of second test user
LMS_TEST_NET_ID2 | < string > | yes | Net ID of second test user
LMS_NOTIFICATIONS_BB_HOST | < url > | yes | Notifications BB base URL
LMS_CORE_BB_CURRENT_HOST | < url > | yes | Core current BB host URL
LMS_CORE_BB_CORE_HOST | < url > | yes | Core BB core host URL
LMS_CORE_BB_HOST | < url > | yes | Core BB host URL
LMS_GROUPS_BB_HOST | < url > | yes | Groups BB host URL
LMS_SERVICE_URL | < url > | yes | URL where this application is being hosted
LMS_SERVICE_ACCOUNT_ID | < string > | yes | ID of Service Account for LMS BB

### Run Application

#### Run locally without Docker

1. Clone the repo (outside GOPATH)

2. Open the terminal and go to the root folder
  
3. Make the project  
```
$ make
...
▶ building executable(s)… 1.9.0 2020-08-13T10:00:00+0300
```

4. Run the executable
```
$ ./bin/content
```

#### Run locally as Docker container

1. Clone the repo (outside GOPATH)

2. Open the terminal and go to the root folder
  
3. Create Docker image  
```
docker build -t content .
```
4. Run as Docker container
```
docker-compose up
```

#### Tools

##### Run tests
```
$ make tests
```

##### Run code coverage tests
```
$ make cover
```

##### Run golint
```
$ make lint
```

##### Run gofmt to check formatting on all source files
```
$ make checkfmt
```

##### Run gofmt to fix formatting on all source files
```
$ make fixfmt
```

##### Cleanup everything
```
$ make clean
```

##### Run help
```
$ make help
```

##### Generate Swagger docs
To run this command, you will need to install [swagger-cli](https://github.com/APIDevTools/swagger-cli)
```
$ make oapi-gen-docs
```


##### Generate models from Swagger docs
To run this command, you will need to install [oapi-codegen](https://github.com/deepmap/oapi-codegen)
```
$ make make oapi-gen-types
```

### Test Application APIs

Verify the service is running as calling the get version API.

#### Call get version API

curl -X GET -i http://localhost/lms/version

Response
```
1.9.0
```

## Contributing
If you would like to contribute to this project, please be sure to read the [Contributing Guidelines](CONTRIBUTING.md), [Code of Conduct](CODE_OF_CONDUCT.md), and [Conventions](CONVENTIONS.md) before beginning.

### Secret Detection
This repository is configured with a [pre-commit](https://pre-commit.com/) hook that runs [Yelp's Detect Secrets](https://github.com/Yelp/detect-secrets). If you intend to contribute directly to this repository, you must install pre-commit on your local machine to ensure that no secrets are pushed accidentally.

```
# Install software 
$ git pull  # Pull in pre-commit configuration & baseline 
$ pip install pre-commit 
$ pre-commit install
```