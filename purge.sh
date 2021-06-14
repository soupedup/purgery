#!/usr/bin/env bash

#MINID=$(date +%s)
#MINID=$(expr ${MINID} - 86400000)-0
#redis-cli -u $REDIS_URL XADD purgery:purge MINID \~ ${MINID} \* url ${1}

curl \
    -X POST \
    -H "Content-Type: application/json" \
    -d "{ \"url\": \"${1}\" }" \
    -u "${API_KEY}": \
    http://localhost:3000/purge
