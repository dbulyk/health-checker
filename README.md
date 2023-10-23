## Description
health-checker is a small application for checking the health of a service. 
It is designed to be used in an environment where the service may be running on a different host than the 
health checker. The health checker will poll the service for CPU and RAM utilization and if it exceeds a certain threshold, 
it will return a 503 error (StatusServiceUnavailable) to the client. In other cases, it returns a 200 (StatusOK) response.


## Flags/Environment Variables
| Flag/Environment Variable | Description                                                                                                        | Default    |
|---------------------------|--------------------------------------------------------------------------------------------------------------------|------------|
| -i / CHECK_INTERVAL       | Interval in seconds by which cpu resource utilization will be polled                                               | 60 seconds |
| -u / THRESHOLD            | Utilization boundary in percent, after which the service will start returning 503 error (StatusServiceUnavailable) | 80         |
| -p / PORT                 | Port on which the health checker will listen for requests                                                          | 8080       |
| -a / ADDRESS              | Address on which the health checker will listen for requests                                                       | localhost  |

Note that if both flags and environment variables are specified, environment variables have higher priority

## Usage
Compile the application with `go build` and run it with `./health-checker`. Specify flags as needed. 
The health checker will listen for requests on the specified port and address.

## Example
`./health-checker -i 10s -u 90 -p 8080 -a localhost`