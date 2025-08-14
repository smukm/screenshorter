#!/bin/bash

for i in {1..10}; do
  echo "Request $i:"
  time curl -X POST http://localhost:8033/api/screen \
    -H "Authorization: Bearer secret" \
    -F "html='<h1>Test $i</h1>'" \
    -F "browser=firefox" \
    -w "\nTime: %{time_total}s\nStatus: %{http_code}\n"
  echo "------------------"
done
