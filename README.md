# Gopulse

A high-performance, concurrent service health checking tool written in Go. This application provides a REST API to monitor the availability and response status of multiple web services simultaneously using a configurable worker pool.

## Features

- Concurrent monitoring: Uses Go routines and worker pools to check multiple services in parallel.
- REST API: Simple HTTP interface to trigger health checks.
- Support for multiple methods: Handles both GET and POST requests.
- Configurable engine: Adjustable timeout settings and worker counts for optimal performance.
- Graceful shutdown: Properly handles OS signals to ensure the server shuts down cleanly.
- Error handling: Detailed reporting for unreachable services or unsupported methods.

## Prerequisites

- Go 1.25.0 or higher

## Getting Started

### Installation

1. Clone this repository to your local machine.
2. Navigate to the project directory:
   ```bash
   cd gopulse
   ```
3. Install the necessary dependencies:
   ```bash
   go mod tidy
   ```

### Running the Application

To start the health checker server:

```bash
go run .
```

The server will start listening on port 8080 by default. You should see a message:
`Health Server started at 8080`

## API Reference

### Check Service Health

Triggers a health check for a list of URLs.

- **URL**: `/health-checker`
- **Method**: `POST`
- **Headers**: `Content-Type: application/json`

#### Request Body

The request body should contain an array of service objects, each specifying the URL and the HTTP method to use.

```json
{
  "urls": [
    {
      "url": "https://www.google.com",
      "method_type": "GET"
    },
    {
      "url": "https://api.example.com/v1/status",
      "method_type": "POST",
      "body": {
        "key": "value"
      }
    }
  ]
}
```

#### Successful Response

Returns the overall status and individual results from the workers.

```json
{
  "status": "Success",
  "message": "Health Checking Completed , Total URLs checked : 2",
  "meta": [
    "Worker 0 says: [OK] https://www.google.com: 200",
    "Worker 1 says: [OK] https://api.example.com/v1/status: 200"
  ]
}
```

## Technical Details

- **Concurrency Model**: The application implements a producer-consumer pattern using Go channels and the `golang.org/x/sync/errgroup` package.
- **Worker Pool**: A "Doctor" object manages a pool of workers (default: 5 workers) that process health check jobs from a shared channel.
- **Timeout Management**: Each health check request has a configurable timeout (default: 10 seconds) to prevent hanging connections from blocking workers.
