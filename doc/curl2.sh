#!/bin/bash

for i in {1..100}; do
  curl -X POST http://localhost:8033/api/screen \
    -H "Authorization: Bearer 12345" \
    -F "html=<h1>Test $i</h1>" \
    -F "browser=firefox" \
    -o /dev/null -s &
done
wait