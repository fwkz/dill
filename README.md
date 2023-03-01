# dill
Cloud ready L4 TCP proxy with first-class support for dynamic listeners.

Exposing dynamic backends on the static frontend ports is the bread-and-butter of any modern proxy. Load balancing multiple dynamic backends from one ingress point using on-demand opened ports is something that, for a good reason as it might poise certain security concerns, is not that simple. But when you exactly know what you are doing you are pretty much on your own.

## Table of Contents
- [dill](#dill)
  - [Table of Contents](#table-of-contents)
  - [Installation](#installation)
    - [Pre-build binaries](#pre-build-binaries)
    - [Build from sources](#build-from-sources)
    - [Docker](#docker)
  - [Getting Started](#getting-started)
  - [Routing](#routing)
    - [Providers](#providers)
      - [File](#file)
      - [HTTP](#http)
      - [Consul](#consul)
    - [Load balancing](#load-balancing)
    - [Schema](#schema)
      - [name](#name)
      - [listener](#listener)
      - [backends](#backends)
      - [proxy _(optional)_](#proxy-optional)
    - [Proxying](#proxying)
  - [Configuration](#configuration)
    - [Values](#values)
      - [listeners.allowed `map`](#listenersallowed-map)
      - [listeners.port\_min `integer`](#listenersport_min-integer)
      - [listeners.port\_max `integer`](#listenersport_max-integer)
      - [peek.listener `string`](#peeklistener-string)
      - [runtime.gomaxprocs `integer`](#runtimegomaxprocs-integer)
      - [routing.file.path `string`](#routingfilepath-string)
      - [routing.file.watch `bool`](#routingfilewatch-bool)
      - [routing.http.endpoint `string`](#routinghttpendpoint-string)
      - [routing.http.poll\_interval `duration`](#routinghttppoll_interval-duration)
      - [routing.http.poll\_timeout `duration`](#routinghttppoll_timeout-duration)
      - [routing.consul.address `string`](#routingconsuladdress-string)
    - [Formats](#formats)
      - [TOML](#toml)
      - [YAML](#yaml)
      - [JSON](#json)
      - [Environment variables](#environment-variables)
  - [Project status](#project-status)
  - [Alternatives](#alternatives)

## Installation
### Pre-build binaries
You can find pre-built binaries in [Releases](https://github.com/fwkz/dill/releases).
### Build from sources
```
$ make build
```
Compiled binary will be available inside `dist/` directory. 
### Docker
```
$ docker pull fwkz/dill
```
or build the image yourself from the sources
```
make image
```
## Getting Started
First of all you have to define how you want to route the incoming traffic. `dill` provides multiple [routing providers](#providers ) that allow you to apply live changes to the routing configuration. By far simplest approach is [routing.file](#file).
```toml
# /etc/dill/config.toml
[routing.file]
  path = "/etc/dill/routing.toml"
  watch = true 
```
```toml
# /etc/dill/routing.toml
[[services]]
  name = "foobar"
  listener = "any:1234"
  backends = ["192.168.10.1:5050"]
``` 
```shell
$ dill -config /etc/dill/config.toml
```
And that's it! `dill` will bind service running on `192.168.10.1:5050` to `0.0.0.0:1234` on the host running the `dill`. Make sure to [read more](#listenersallowed-map) about interface labels, e.g., `any`, `local`


## Routing
### Providers
`dill` offers multiple routing providers that allow you to apply live changes to the routing configuration.
- [File](#file)
- [HTTP](#http)
- [Consul](#consul)
#### File
It is the simplest way of defining routing. All routing logic is being kept in a separate [config file](#schema). By setting `routing.file.watch = true` you can also subscribe to changes made to the routing configuration file which would give you the full power of dill's dynamic routing capabilities.
```toml
[routing.file]
  path = "/etc/dill/routing.toml"
  watch = true 
```
#### HTTP
`dill` can poll the HTTP endpoint for its routing configuration with a predefined time interval. Fetched data should be compliant with [routing configuration schema](#schema) and it will be parsed based on the response `Content-Type` header.
```toml
[routing.http]
  endpoint = "http://127.0.0.1:8000/config/routing.json"
  poll_interval = "5s"
  poll_timeout = "5s"
```
#### Consul
`dill` can build its routing table based on services registered in `Consul`. All you need to do in order to expose Consul registered service in `dill` instace is to add appropriate tags.
* `dill` tag registers service and its updates with `dill` instance.
* `dill.listener` binds, based on predefined listeners declared by [`listeners.allowed`](#listenersallowed-map), service to specific address and port.
```json
{
  "service": {
    "tags": [
      "dill",
      "dill.listener=local:5555",
    ],
  }
}
```
In order to pass traffic via [proxy](#proxying) make sure to add `dill.proxy` tag:
```json
{
  "service": {
    "tags": [
      "dill",
      "dill.listener=local:5555",
      "dill.proxy=socks5://admin:password@192.168.10.11:1080"
    ],
  }
}
````
### Load balancing
`dill` distributes load across the backends using _round-robin_ strategy

### Schema
The routing configuration should be compliant with following schema: 
```toml
# /etc/dill/routing.toml
[[services]]
  name = "foo"
  listener = "local:1234"
  backends = ["127.0.0.1:5050"]
  proxy = "socks5://user:pass@127.0.0.1:1080"  # optional

[[services]]
  name = "bar"
  listener = "any:4444"
  backends = ["127.0.0.1:4000", "127.0.0.1:4001"]
```
or equivalent in different format e.g. `JSON`, `YAML`:
```json
{
  "services": [
    {
      "name": "foo",
      "listener": "local:1234",
      "backends": [
        "127.0.0.1:5050"
      ],
      "proxy": "socks5://user:pass@127.0.0.1:1080"
    },
    {
      "name": "bar",
      "listener": "any:4444",
      "backends": [
        "127.0.0.1:4000",
        "127.0.0.1:4001"
      ]
    }
  ]
}
```
```yaml
services:
  - name: foo
    listener: local:1234
    backends:
      - 127.0.0.1:5050
    proxy: socks5://user:pass@127.0.0.1:1080
  - name: bar
    listener: any:4444
    backends:
      - 127.0.0.1:4000
      - 127.0.0.1:4001
```
#### name
Name of the service.
#### listener
Listener that binds, based on predefined list declared by [`listeners.allowed`](#listenersallowed-map), backend to specific address and port.
#### backends
List of backend services that will be load balanced.
#### proxy _(optional)_
[Proxy](#proxying) address if you want to tunnel the traffic.
### Proxying
`dill` is capable of tunneling traffic to backend services using SOCKS proxy.
```toml
# /etc/dill/routing.toml
[[services]]
  name = "foobar"
  listener = "any:4444"
  backends = ["192.168.10.11:4001"]
  proxy = "socks5://user:password@192.168.10.10:1080"
```
```text
     incoming      ┌───────────┐        ┌─────────────┐         ┌───────────────┐
─────connection────►4444  dill ├────────►1080  SOCKS  ├─────────►4001  Backend  │
                   └───────────┘        └─────────────┘         └───────────────┘
                                         192.168.10.10            192.168.10.11
```

## Configuration
`dill` already comes with sane defaults but you can adjust its behaviour providing configuration file 
```bash
$ dill -config config.toml
``` 
or use environment variables
```bash
$ export DILL_CONSUL_ADDRESS="http://127.0.0.1:8500"
$ DILL_LISTENERS_PORT_MIN=1234 dill
``` 

### Values
#### listeners.allowed `map`
Interface addresses that are allowed to be bind to by upstream services. Address labels (keys in the map) are opaque for `dill`. 

Imagine that a machine hosting `dill` has two interfaces, one is internal (192.168.10.10) and the other is external (12.42.22.65). You might want to use the following setup 
```toml
[listeners.allowed]
internal = "192.168.10.10"
public = "12.42.22.65"
```
with such configuration, upstream services that want to be accessible on `12.42.22.65:5555` can use the `public` listener in Consul tags `dill.listener=public:5555`. 

_default: `{"local": "127.0.0.1", "any": "0.0.0.0"}`_
#### listeners.port_min `integer`
Minimal port value at which it will be allowed to expose upstream services. Backends requesting to be exposed on lower ports will be dropped from routing.

_default: `1024`_
#### listeners.port_max `integer`
Maximum port value at which it will be allowed to expose upstream services. Backends requesting to be exposed on higher ports will be dropped from routing.

_default: `49151`_
#### peek.listener `string`
Address on which `Peek` will be exposed. `Peek` is a TCP debug server spawned alongside the `dill`. Connecting to it will return the current state of the routing table. By default `Peek` is turned off.

_default: `""`_
```
$ nc 127.0.0.1 2323
0.0.0.0:4444
  ├ round_robin
  ├──➤ 192.168.10.17:1234
  ├──➤ 192.168.10.23:2042
0.0.0.0:8088
  ├ round_robin
  ├──➤ 192.168.10.11:5728
  ├──➤ 192.168.65.87:5942
```
#### runtime.gomaxprocs `integer`
Value of Go's `runtime.GOMAXPROCS()`

_default: equals to `runtime.NumCPU()`_

#### routing.file.path `string`
Location of [routing configuration file](#schema).
#### routing.file.watch `bool`
Subscribe to changes made to the routing configuration file which would give you the full power of dill's dynamic routing capabilities.

_default: `true`_
#### routing.http.endpoint `string`
Endpoint which [http provider](#http) will poll for routing configuration
#### routing.http.poll_interval `duration`
How often [http provider](#http) will poll [endpoint](#routinghttpendpoint-string) for routing configuration

_default: `5s`_
#### routing.http.poll_timeout `duration`
Maximum time  [http provider](#http) will wait when fetching routing configuration

_default: `5s`_
#### routing.consul.address `string`
Consul address from which `dill` will fetch the updates and build the routing table.
### Formats
Configuration is powered by [Viper](https://github.com/spf13/viper) so it's possible to use format that suits you best.

> reading from JSON, TOML, YAML, HCL, envfile and Java properties config files

`dill` uses the following precedence order:
  * environment variable
  * config file
  * default value


#### TOML
```toml
[listeners]
port_min = 1024
port_max = 49151

[listeners.allowed]
local = "127.0.0.1"
any = "0.0.0.0"

[routing.consul]
address = "http://127.0.0.1:8500"

[peek]
listener = "127.0.0.1:4141"

[runtime]
gomaxprocs = 4
```
#### YAML
```yaml
listeners:
  port_min: 1024 
  port_max: 49151
  allowed:
    local: "127.0.0.1"
    any: "0.0.0.0"

routing:
  http:
    endpoint: "http://127.0.0.1:8000/config/routing.json"
    poll_interval: "5s"
    poll_timeout: "5s"

peek:
  listener: "127.0.0.1:4141"

runtime:
  gomaxprocs: 4
```
#### JSON
```json
{
  "listeners": {
    "port_min": 1024,
    "port_max": 49151,
    "allowed": {
      "local": "127.0.0.1",
      "any": "0.0.0.0"
    }
  },
  "routing": {
    "file": {
      "path": "/Users/fwkz/Devel/dill/configs/routing.toml",
      "watch": true
    }
  },
  "peek": {
    "listener": "127.0.0.1:4141"
  },
  "runtime": {
    "gomaxprocs": 4
  }
}
```
#### Environment variables
Variables should be prefixed with `DILL` and delimited with underscore e.g. `consul.address` becomes `DILL_CONSUL_ADDRESS`. 
```bash
$ export DILL_CONSUL_ADDRESS="http://127.0.0.1:8500"
$ DILL_LISTENERS_PORT_MIN=1234 dill
``` 

## Project status
Concept of dynamic listeners is experimental and should be used with 
responsibility. There might be some breaking changes in the future. 

## Alternatives
* [fabio](https://github.com/fabiolb/fabio)
* [traefik](https://github.com/traefik/traefik)
* [RSOCKS](https://github.com/tonyseek/rsocks)
