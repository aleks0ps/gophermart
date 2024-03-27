#!/bin/sh -x
# valid numbers
NUMBERS=$(cat numbers.txt)
URL="localhost:8080/api/orders"
i=0
for order in $NUMBERS; do
  i=$((i+1))
  # random price
  price=$(od -vAn -N2 -tu2 < /dev/urandom | tr -d ' ')
  curl -X POST -H "Content-Type: application/json" -d "{
    \"order\": \"$order\",
    \"goods\": [{\"description\":\"item${i}\", \"price\": $price }]
  }" $URL 
done
