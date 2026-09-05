# Go Load Balancer

A high-performance, multi-service load balancer written in Go, supporting Layer 4 (TCP) and Layer 7 (HTTP) proxying. It routes traffic across independently managed services, where each service can expose multiple backend replicas and use its own load-balancing strategy. The system provides application-aware routing, health-aware backend selection, concurrent traffic distribution, and an extensible architecture for adding new routing and balancing policies.

## Features

### Layer 4

- **TCP Proxying:** Accepts incoming TCP connections, selects a healthy backend,
  establishes a connection to the selected backend, and proxies traffic in both
  directions without inspecting the application payload.

- **5-Tuple Load Balancing:** Uses source IP, destination IP, source port,
  destination port, and protocol information for connection-level backend
  selection.

### Layer 7

- **HTTP Reverse Proxy:** Inspects HTTP requests and forwards them to the
  appropriate backend service.

- **Application-Aware Routing:** Supports routing based on HTTP request
  information such as paths and other application-level attributes.

- **Load Balancing Strategies:** Supports Round-Robin and Weighted
  Round-Robin for distributing requests across backend instances.

- **TLS Termination:** Supports terminating TLS at the proxy layer before
  forwarding traffic to backend services.

### Reliability & Operations

- **Multi-Tier Architecture:** Supports deploying an L4 load balancer in front
  of multiple L7 instances, allowing the application layer to scale
  independently.

- **Active Health Checking:** Continuously monitors backend availability and
  removes unhealthy backends from the active pool.

- **Failure Handling:** Detects backend failures and supports retrying requests
  against healthy backends where appropriate.

- **Header Management:** Manages HTTP forwarding headers such as
  `X-Forwarded-For` to preserve client information across proxy layers.

- **Concurrent Connection Handling:** Uses Go's concurrency primitives to
  handle multiple TCP connections and HTTP requests concurrently.

- **Configurable Backend Pools:** Supports multiple backend instances and
  configurable load-balancing strategies.

## Architecture

The load balancer can be deployed as a multi-tier system where an L4 layer distributes TCP connections across multiple L7 load balancers. The L7 layer then handles HTTP routing and distributes requests across backend services.

```text
                    Client
                       |
                       v
              +-----------------+
              |  L7 Load Balancer|
              +--------+--------+
                       |
          +------------+------------+
          |            |            |
          v            v            v
       Users        Payments       Orders
         |              |              |
     Round Robin   Least Conn.   Weighted RR
          |            |            |
          v            v            v
       [Pool]        [Pool]        [Pool]
        / | \         / | \         / | \
       U1 U2 U3      P1 P2 P3      O1 O2 O3
```

## Configuration

The load balancer is configured using a YAML configuration file.

```yaml
mode: l7

services:
  - name: users
    matcher: /users
    strategy: RoundRobin
    replicas:
      - http://localhost:8081
      - http://localhost:8082
```

## Usage

```bash
go run .
```

<!-- *** Building a Distributed Load Balancer *** -->