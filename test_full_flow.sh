#!/usr/bin/env bash

BASE="http://localhost:8080"
set -e

echo "=============================="
echo "1) Health check"
echo "=============================="
curl -s "$BASE/ping" | jq

echo "=============================="
echo "2) Get categories"
echo "=============================="
CATEGORIES=$(curl -s "$BASE/categories")
echo "$CATEGORIES" | jq
CATEGORY_ID=$(echo "$CATEGORIES" | jq -r '.data[0].id')
echo "Using category_id=$CATEGORY_ID"

echo "=============================="
echo "3) Create user"
echo "=============================="
UNIQUE=$(date +%s)

USER_RESPONSE=$(curl -s -X POST "$BASE/users" \
  -H "Content-Type: application/json" \
  -d "{
    \"name\": \"Test User $UNIQUE\",
    \"phone\": \"+7999$UNIQUE\",
    \"email\": \"testuser$UNIQUE@example.com\",
    \"password_hash\": \"fake_hash\"
  }")

echo "$USER_RESPONSE" | jq
USER_ID=$(echo "$USER_RESPONSE" | jq -r '.data.id')

if [ "$USER_ID" = "null" ] || [ -z "$USER_ID" ]; then
  echo "Failed to create user"
  exit 1
fi

echo "Created user_id=$USER_ID"

echo "=============================="
echo "4) Create menu item"
echo "=============================="
MENU_RESPONSE=$(curl -s -X POST "$BASE/menu" \
  -H "Content-Type: application/json" \
  -d "{
    \"name\": \"Test Burger $UNIQUE\",
    \"description\": \"Created by test script\",
    \"price\": \"500.00\",
    \"category_id\": $CATEGORY_ID,
    \"is_available\": true
  }")

echo "$MENU_RESPONSE" | jq
MENU_ITEM_ID=$(echo "$MENU_RESPONSE" | jq -r '.data.id')
echo "Created menu_item_id=$MENU_ITEM_ID"

echo "=============================="
echo "5) Place order"
echo "=============================="
ORDER_RESPONSE=$(curl -s -X POST "$BASE/orders" \
  -H "Content-Type: application/json" \
  -d "{
    \"user_id\": $USER_ID,
    \"items\": [
      {
        \"menu_item_id\": $MENU_ITEM_ID,
        \"quantity\": 2
      }
    ],
    \"promo_code\": \"\",
    \"delivery_address\": \"Innopolis test address\",
    \"notes\": \"Created by full flow script\"
  }")

echo "$ORDER_RESPONSE" | jq
ORDER_ID=$(echo "$ORDER_RESPONSE" | jq -r '.data.order.id')

if [ "$ORDER_ID" = "null" ] || [ -z "$ORDER_ID" ]; then
  echo "Failed to create order"
  exit 1
fi

echo "Created order_id=$ORDER_ID"

echo "=============================="
echo "6) Test order GET endpoints"
echo "=============================="
curl -s "$BASE/orders" | jq
curl -s "$BASE/orders/pending" | jq
curl -s "$BASE/orders/$ORDER_ID" | jq
curl -s "$BASE/orders/$ORDER_ID/items" | jq
curl -s "$BASE/orders/$ORDER_ID/track" | jq
curl -s "$BASE/users/$USER_ID/orders" | jq

echo "=============================="
echo "7) Update order status flow"
echo "=============================="
curl -s -X PATCH "$BASE/orders/$ORDER_ID/status" \
  -H "Content-Type: application/json" \
  -d '{"status": "confirmed"}' | jq

curl -s -X PATCH "$BASE/orders/$ORDER_ID/status" \
  -H "Content-Type: application/json" \
  -d '{"status": "preparing"}' | jq

curl -s -X PATCH "$BASE/orders/$ORDER_ID/status" \
  -H "Content-Type: application/json" \
  -d '{"status": "ready"}' | jq

echo "=============================="
echo "8) Courier endpoints"
echo "=============================="
COURIERS=$(curl -s "$BASE/couriers")
echo "$COURIERS" | jq
COURIER_ID=$(echo "$COURIERS" | jq -r '.data[0].id')

if [ "$COURIER_ID" = "null" ] || [ -z "$COURIER_ID" ]; then
  echo "No couriers found in database."
  exit 1
fi

echo "Using courier_id=$COURIER_ID"

curl -s "$BASE/couriers/available" | jq
curl -s "$BASE/couriers/$COURIER_ID" | jq

curl -s -X PATCH "$BASE/couriers/$COURIER_ID/status" \
  -H "Content-Type: application/json" \
  -d '{"status": "available"}' | jq

echo "=============================="
echo "9) Assign courier"
echo "=============================="
curl -s -X PATCH "$BASE/orders/$ORDER_ID/courier" \
  -H "Content-Type: application/json" \
  -d "{
    \"courier_id\": $COURIER_ID
  }" | jq

echo "=============================="
echo "10) Final tracking"
echo "=============================="
curl -s "$BASE/orders/$ORDER_ID/track" | jq

echo "=============================="
echo "11) Create another order then cancel it"
echo "=============================="
CANCEL_ORDER_RESPONSE=$(curl -s -X POST "$BASE/orders" \
  -H "Content-Type: application/json" \
  -d "{
    \"user_id\": $USER_ID,
    \"items\": [
      {
        \"menu_item_id\": $MENU_ITEM_ID,
        \"quantity\": 1
      }
    ],
    \"promo_code\": \"\",
    \"delivery_address\": \"Cancel test address\",
    \"notes\": \"This order will be cancelled\"
  }")

echo "$CANCEL_ORDER_RESPONSE" | jq
CANCEL_ORDER_ID=$(echo "$CANCEL_ORDER_RESPONSE" | jq -r '.data.order.id')

echo "Created cancel_order_id=$CANCEL_ORDER_ID"

curl -s -X PATCH "$BASE/orders/$CANCEL_ORDER_ID/cancel" | jq
curl -s "$BASE/orders/$CANCEL_ORDER_ID/track" | jq

echo "=============================="
echo "DONE"
echo "=============================="