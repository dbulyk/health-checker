## Description
health-checker is a small application for checking the health of a service. 
It is designed to be used in an environment where the service may be running on a different host than the 
health checker. The health checker will poll the service for CPU and RAM utilisation and if some of these things exceeds a certain threshold,
it will return a 503 error (StatusServiceUnavailable) to the client. In other cases, it returns a 200 (StatusOK) response.


## Flags/Environment Variables
| Flag/Environment Variable | Description                                                                                                                                              | Default    |
|---------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------|------------|
| -i / CHECK_INTERVAL       | Interval in seconds by which cpu resource utilisation will be polled _(Do not set the interval too large, <br/>otherwise the values will be too smooth)_ | 60 seconds |
| -u / THRESHOLD            | utilisation boundary in percent, after which the service will start returning 503 error _(StatusServiceUnavailable)_                                     | 80         |
| -p / PORT                 | Port on which the health checker will listen for requests                                                                                                | 8080       |
| -a / ADDRESS              | Address on which the health checker will listen for requests                                                                                             | localhost  |
| -d / DEBUG                | If set to true, the health checker will print debug information to stdout _(Do not specify it if you don't need debug messages)_                         | false      |

_Note that if both flags and environment variables are specified, environment variables have higher priority_

## Usage
Compile the application with `go build` or download it from releases and run it with `./health-checker`. Specify flags as needed.

## Handling requests
The health checker will respond to GET requests on the `address:port/check` endpoint.

## Example
For macOS, run `./health-checker -i 10s -u 90 -p 8080 -a localhost -d` or `CHECK_INTERVAL=10s THRESHOLD=90 PORT=8080 ADDRESS=localhost DEBUG=true ./health-checker` </br></br>
For Windows, run `health-checker.exe -i 10s -u 90 -p 8080 -a localhost -d` or `CHECK_INTERVAL=10s THRESHOLD=90 PORT=8080 ADDRESS=localhost DEBUG=true health-checker.exe`

## Libraries used
- [Gopsutil](https://github.com/shirou/gopsutil) for getting CPU and RAM utilisation
