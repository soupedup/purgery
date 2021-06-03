Purgery is a lightweight cache purging service intended to run alongside HTTP cache servers (varnish, nginx, etc).

## The problem

Some popular open source caching proxies don't offer official tools to issue cache purges across multiple instances. If you run caches in multiple
regions, you'll need something like this to ensure you aren't serving stale content to users.

## Proxy support

Purgery only supports Varnish, and only versions that accept BAN verb requests over HTTP.

## Requirements

Purgery requires Redis 6 to run. It uses Redis streams to ensure that purge fails can be reliably delivered and replayed after outages.

The following environment variables must be set in production:

`REDIS_URL`: Your redis 6 connection string
`VARNISH_ADDR`: The varnish server address this instance should target

## Deploying in Fly.io
Optionally, you can set `PROXY_APP_NAME` when deploying on Fly.io to automatically set `VARNISH_ADDR` to the regional sibling.

A sample Fly.io setup is coming soon.
