# dill
Proxy with first-class support for dynamic listeners. 

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
    - [File](#file)
    - [HTTP](#http)
    - [Consul](#consul)
  - [Configuration](#configuration)
    - [Values](#values)
      - [consul.address `string`](#consuladdress-string)
      - [listeners.allowed `map`](#listenersallowed-map)
      - [listeners.port\_min `integer`](#listenersport_min-integer)
      - [listeners.port\_max `integer`](#listenersport_max-integer)
      - [peek.listener `string`](#peeklistener-string)
      - [runtime.gomaxprocs `integer`](#runtimegomaxprocs-integer)
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
First of all we have to define how we want to route the incoming traffic. `dill` provides multiple [routing providers](#routing) that allow you to apply live changes to the routing configuration. By far simplest approach is [routing.file](#file).
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
`dill` offers multiple routing providers that allow you to apply live changes to the routing configuration.
### File
It is the simplest way of defining routing. All routing logic is being kept in a separate config file. By setting `routing.file.watch = true` you can also subscribe to changes made to the routing configuration file which would give you the full power of dill's dynamic routing capabilities.
```toml
[routing.file]
  path = "/etc/dill/routing.toml"
  watch = true 
```
The routing configuration defined by `routing.file.path` should look like this:
```toml
# /etc/dill/routing.toml
[[services]]
  name = "foo"
  listener = "local:1234"
  backends = ["127.0.0.1:5050"]

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
      ]
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
  - name: bar
    listener: any:4444
    backends:
      - 127.0.0.1:4000
      - 127.0.0.1:4001
```
### HTTP
`dill` can poll the HTTP endpoint for its routing configuration with a predefined time interval. Fetched data should have the same format as the routing config file defined by `routing.file.path` and will be parsed based on the response `Content-Type` header.
```toml
[routing.http]
  endpoint = "http://127.0.0.1:8000/config/routing.json"
  poll_interval = "5s"
  poll_timeout = "5s"
```
### Consul
`dill` can build its routing table based on services registered in `Consul`. All you need to do in order to expose Consul registered service in `dill` instace is to add appropriate tags.
* `dill` tag registers service and its updates with `dill` instance.
* `dill.listener` binds, based on predefined listeners declared by `listeners.allowed`, service to specific address and port.
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
#### consul.address `string`
Consul address from which `dill` will fetch the updates and build the routing table.`

_default: `http://127.0.0.1:8500`_
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
