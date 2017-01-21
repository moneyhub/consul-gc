# consul-gc

[![Build Status](https://travis-ci.org/momentumft/consul-gc.svg?branch=master)](https://travis-ci.org/momentumft/consul-gc) [![Docker Repository on Quay](https://quay.io/repository/momentumft/consul-gc/status "Docker Repository on Quay")](https://quay.io/repository/momentumft/consul-gc)

`consul-gc` is a small daemon service that helps cleanup failed/lost Consul servers when running a Consul cluster within an AWS Auto Scaling Group. The need for `consul-gc` came from our attempts at running Consul in a kubernetes Deployment/StatefulSet but it should work in any EC2 cluster within an Auto Scaling Group.

Currently `consul-gc` takes a very opinionated view of your Consul cluster and assumes that the expected number of Consul servers is the same as the desired instance count of your Auto Scaling Group. Any additional Consul servers in the `failed` Serf state will be removed via the `/v1/agent/force-leave` api.

## Configuration

### Comand Line Options

```
Usage of consul-gc:
  -interval int
        interval between checking members (default 60)
  -v	show version
```

### Environment Variables

Use the following environment variables for the consul agent connection:

- `CONSUL_HTTP_ADDR`
- `CONSUL_HTTP_TOKEN`
- `CONSUL_HTTP_AUTH`
- `CONSUL_HTTP_SSL`
- `CONSUL_HTTP_SSL_VERIFY`

These varaibles are used directly by the Consul Go api, so we recommend checking the consul [docs](https://godoc.org/github.com/hashicorp/consul/api#pkg-constants) for more information.
