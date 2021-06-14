#!/usr/bin/env bash

MINID=$(date +%s)
MINID=$(expr ${MINID} - 86400000)-0
REDIS_URL=redis://localhost:6380
redis-cli -u $REDIS_URL XADD purgery:purge MINID \~ ${MINID} \* url https://www.example.com
