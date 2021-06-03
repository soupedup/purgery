#!/usr/bin/env bash

PURGERY_ID=${FLY_ALLOC_ID}
VARNISH_ADDR=${FLY_REGION}.${PROXY_APP_NAME}.internal:80
/usr/local/bin/purgery
