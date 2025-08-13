#!/bin/bash

#!/bin/bash

# Проверяем существование файла
if [ ! -f "post_data.txt" ]; then
  echo "Создаю post_data.txt..."
  cat << 'EOF' > post_data.txt
--WebKitFormBoundary7MA4YWxkTrZu0gW
Content-Disposition: form-data; name="html"

<h1>Load Test</h1>
--WebKitFormBoundary7MA4YWxkTrZu0gW
Content-Disposition: form-data; name="browser"

firefox
--WebKitFormBoundary7MA4YWxkTrZu0gW--
EOF
fi

# Запускаем тест
ab -n 100 -c 100 \
  -T "multipart/form-data; boundary=WebKitFormBoundary7MA4YWxkTrZu0gW" \
  -p post_data.txt \
  -H "Authorization: Bearer 12345" \
  http://localhost:8033/api/screen