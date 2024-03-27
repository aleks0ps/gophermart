#!/bin/sh

NUMBERS=$(cat numbers.txt)
URL="localhost:8080/api/orders"
for order in $NUMBERS; do
  curl $URL/$order
done
