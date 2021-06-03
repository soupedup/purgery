![Automated tests](https://github.com/soupedup/purgery/actions/workflows/test.yml/badge.svg)

Purgery is a lightweight cache purging service intended to run alongside caching reverse proxies (varnish, nginx, etc).

## The problem

Some popular open source caching proxies don't offer official tools to issue cache purges across multiple instances. If you run caches in multiple regions, you'll need something like this to ensure you aren't serving stale content to users.

## Requirements

Purgery requires a Redis 6 instance to distribute requests. [Redis streams](https://redis.io/topics/streams-intro) ensure that purge fails can be reliably delivered, and replayed from a checkpoint after outages.

## Deployment

Purgery may be run as its own service, but you should ensure that it runs one instance per caching proxy instance, ideally in the same deployment region.

The following environment variables must be set in production:

`REDIS_URL`: Your redis 6 connection string
`VARNISH_ADDR`: The varnish server address this instance should target

Purgery only supports Varnish, and only versions that accept BAN verb requests over HTTP.

[We provide a Varnish image](https://github.com/soupedup/varnish) with:

* Default configuration supporting BAN requests over HTTP
* An option to run Purgery in the same container alongside Varnish

## Issuing cache purge requests

Purging is done by a single [XADD](https://redis.io/commands/xadd) command sent to the redis `purgery:purge` key. See [purge.sh](https://github.com/soupedup/purgery/blob/main/purge.sh). `XADD` takes two arguments:

`MINID`: Timestamp in the past at which previous entries should be truncated. This is used as a simple mechanism to keep the stream from filling up indefinitely.
`url`: Full URL to be purged. For now, the default Varnish configuration only supports [purging entire domains](https://github.com/soupedup/varnish/blob/main/default.vcl#L25). The path is ignored.

## Deploying in Fly.io

Optionally, you can set `PROXY_APP_NAME` when deploying on Fly.io to automatically set `VARNISH_ADDR` to the instance of that Fly app in the same region as Purgery.

A sample Fly.io setup is coming soon.

## Development

An example stack can be run with `docker-compose up` and using the accompanying "