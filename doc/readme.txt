curl -X POST http://localhost:8033/api/screen \
  -H "Authorization: Bearer 12345" \
  -F "html='<h1>Test</h1>'" \
  -F "browser=firefox"