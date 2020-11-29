# dyntcp
Dynamic Listener TCP Proxy integrated with Hashicorp Consul.
Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed in lorem iaculis sem pretium pharetra. Pellentesque nulla ex, facilisis vel urna sed, luctus gravida nibh. Cras quis nibh mi. Maecenas libero massa, suscipit sed varius non, semper ac ipsum. Etiam eget velit vulputate, tincidunt sem posuere, eleifend odio. Cras commodo.

## Motivation
Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed in lorem iaculis sem pretium pharetra. Pellentesque nulla ex, facilisis vel urna sed, luctus gravida nibh. Cras quis nibh mi. Maecenas libero massa, suscipit sed varius non, semper ac ipsum. Etiam eget velit vulputate, tincidunt sem posuere, eleifend odio. Cras commodo. 

## Configuration
`dyntcp` already comes with sane defaults but you can adjust its behaviour providing configuration file or use environment variables. 
```bash
$ dyntcp -c config.toml
``` 

### Values
#### consul.address `string`
Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed in lorem iaculis sem pretium pharetra. Pellentesque nulla ex, facilisis vel urna sed, luctus gravida nibh. Cras quis nibh mi. Maecenas libero massa, suscipit sed varius non, semper ac ipsum. Etiam eget velit vulputate, tincidunt sem posuere, eleifend odio. Cras commodo. 
#### listeners.allowed `map`
Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed in lorem iaculis sem pretium pharetra. Pellentesque nulla ex, facilisis vel urna sed, luctus gravida nibh. Cras quis nibh mi. Maecenas libero massa, suscipit sed varius non, semper ac ipsum. Etiam eget velit vulputate, tincidunt sem posuere, eleifend odio. Cras commodo. 
#### listeners.port_min `integer`
Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed in lorem iaculis sem pretium pharetra. Pellentesque nulla ex, facilisis vel urna sed, luctus gravida nibh. Cras quis nibh mi. Maecenas libero massa, suscipit sed varius non, semper ac ipsum. Etiam eget velit vulputate, tincidunt sem posuere, eleifend odio. Cras commodo. 
#### listeners.port_max `integer`
Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed in lorem iaculis sem pretium pharetra. Pellentesque nulla ex, facilisis vel urna sed, luctus gravida nibh. Cras quis nibh mi. Maecenas libero massa, suscipit sed varius non, semper ac ipsum. Etiam eget velit vulputate, tincidunt sem posuere, eleifend odio. Cras commodo. 
#### peek.listener `string`
Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed in lorem iaculis sem pretium pharetra. Pellentesque nulla ex, facilisis vel urna sed, luctus gravida nibh. Cras quis nibh mi. Maecenas libero massa, suscipit sed varius non, semper ac ipsum. Etiam eget velit vulputate, tincidunt sem posuere, eleifend odio. Cras commodo. 
#### runtime.gomaxprocs `integer`
Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed in lorem iaculis sem pretium pharetra. Pellentesque nulla ex, facilisis vel urna sed, luctus gravida nibh. Cras quis nibh mi. Maecenas libero massa, suscipit sed varius non, semper ac ipsum. Etiam eget velit vulputate, tincidunt sem posuere, eleifend odio. Cras commodo. 
### Formats
Configuration is powered by [Viper](https://github.com/spf13/viper) so it's possible to use format that suits you best.

> reading from JSON, TOML, YAML, HCL, envfile and Java properties config files

`dyntcp` uses the following precedence order:
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

[consul]
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

consul:
  address: "http://127.0.0.1:8500"

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
  "consul": {
    "address": "http://127.0.0.1:8500"
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
Variables should be prefixed with `DYNTCP` and delimited with underscore e.g. `consul.address` becomes `DYNTCP_CONSUL_ADDRESS`. 
```
DYNTCP_CONSUL_ADDRESS=http://127.0.0.1
```

## Project status
Concept of dynamic listeners is experimental and should be used with 
responsibility. There might be some breaking changes in the future. 

## Acknowledgments
* fabio
