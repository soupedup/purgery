#!/usr/bin/env bash

MINID=$(date +%s%3N)
MINID=$(expr ${MINID} - 86400000)-0

redis-cli -u $REDIS_URL XADD purgery:purge MINID \~ ${MINID} \* url z   