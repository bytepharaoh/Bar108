#!/usr/bin/env bash

BASE_URL="http://localhost:8080"

CUSTOMER1_EMAIL="customer1@bar108.com"
CUSTOMER2_EMAIL="customer2@bar108.com"
ADMIN_EMAIL="admin@bar108.com"
PASSWORD="secret123"

PASS=0
FAIL=0

report() {
  local name="$1"
  local expected="$2"
  local actual="$3"

  if [ "$expected" = "$actual" ]; then
    echo "✅ PASS: $name | status=$actual"
    PASS=$((PASS + 1))
  else
    echo "❌ FAIL: $name | expected=$expected got=$actual"
    FAIL=$((FAIL + 1))
  fi
}

register_user() {
  local name="$1"
  local phone="$2"
  local email="$3"

  curl -s -X POST "$BASE_URL/auth/register" \
    -H "Content-Type: application/json" \
    -d "{\"name\":\"$name\",\"phone\":\"$phone\",\"email\":\"$email\",\"password\":\"$PASSWORD\"}" > /dev/null
}

login_user() {
  local email="$1"

  curl -s -X POST "$BASE_URL/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$email\",\"password\":\"$PASSWORD\"}" \
    | sed -n 's/.*"token":"\([^"]*\)".*/\1/p'
}

status_code() {
  curl -s -o /tmp/bar108_response.json -w "%{http_code}" "$@"
}

echo "=============================="
echo "Bar108 Auth/Orders Test Script"
echo "=============================="

echo
echo "1) Creating users..."
register_user "Customer One" "+79000000001" "$CUSTOMER1_EMAIL"
register_user "Customer Two" "+79000000002" "$CUSTOMER2_EMAIL"
register_user "Admin User" "+79000000003" "$ADMIN_EMAIL"

echo "Users created or already exist."

echo
echo "2) Promote admin manually if not already admin:"
echo "Run this in DB if needed:"
echo "UPDATE users SET role = 'admin' WHERE email = '$ADMIN_EMAIL';"
echo

read -p "Press Enter after promoting admin in DB..."

echo
echo "3) Logging in and saving tokens..."

CUSTOMER1_TOKEN=$(login_user "$CUSTOMER1_EMAIL")
CUSTOMER2_TOKEN=$(login_user "$CUSTOMER2_EMAIL")
ADMIN_TOKEN=$(login_user "$ADMIN_EMAIL")

echo "$CUSTOMER1_TOKEN" > customer1.token
echo "$CUSTOMER2_TOKEN" > customer2.token
echo "$ADMIN_TOKEN" > admin.token

echo "Tokens saved:"
echo "- customer1.token"
echo "- customer2.token"
echo "- admin.token"

echo
echo "4) Extracting user IDs from login responses..."

CUSTOMER1_ID=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$CUSTOMER1_EMAIL\",\"password\":\"$PASSWORD\"}" \
  | sed -n 's/.*"id":\([0-9]*\).*/\1/p')

CUSTOMER2_ID=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$CUSTOMER2_EMAIL\",\"password\":\"$PASSWORD\"}" \
  | sed -n 's/.*"id":\([0-9]*\).*/\1/p')

echo "Customer 1 ID: $CUSTOMER1_ID"
echo "Customer 2 ID: $CUSTOMER2_ID"

echo
echo "5) Running test cases..."
echo

# Public route
CODE=$(status_code "$BASE_URL/menu")
report "Public GET /menu should work without token" "200" "$CODE"

# Customer should NOT access admin all orders
CODE=$(status_code "$BASE_URL/orders" \
  -H "Authorization: Bearer $CUSTOMER1_TOKEN")
report "Customer GET /orders should be forbidden" "403" "$CODE"

# Admin should access all orders
CODE=$(status_code "$BASE_URL/orders" \
  -H "Authorization: Bearer $ADMIN_TOKEN")
report "Admin GET /orders should work" "200" "$CODE"

# Customer 1 creates order
CODE=$(status_code -X POST "$BASE_URL/orders" \
  -H "Authorization: Bearer $CUSTOMER1_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"user_id\":$CUSTOMER1_ID,\"items\":[{\"menu_item_id\":1,\"quantity\":1}]}")
report "Customer 1 POST /orders should create order" "201" "$CODE"

# Customer 1 views his own orders
CODE=$(status_code "$BASE_URL/users/$CUSTOMER1_ID/orders" \
  -H "Authorization: Bearer $CUSTOMER1_TOKEN")
report "Customer 1 GET own /users/:id/orders should work" "200" "$CODE"

# Customer 1 tries to view Customer 2 orders
CODE=$(status_code "$BASE_URL/users/$CUSTOMER2_ID/orders" \
  -H "Authorization: Bearer $CUSTOMER1_TOKEN")
report "Customer 1 GET another user's orders should be forbidden" "403" "$CODE"

# Customer 2 tries to view Customer 1 orders
CODE=$(status_code "$BASE_URL/users/$CUSTOMER1_ID/orders" \
  -H "Authorization: Bearer $CUSTOMER2_TOKEN")
report "Customer 2 GET another user's orders should be forbidden" "403" "$CODE"

# Admin views customer orders
CODE=$(status_code "$BASE_URL/users/$CUSTOMER1_ID/orders" \
  -H "Authorization: Bearer $ADMIN_TOKEN")
report "Admin GET /users/:id/orders should work" "200" "$CODE"

# No token
CODE=$(status_code "$BASE_URL/users/$CUSTOMER1_ID/orders")
report "No token should be unauthorized" "401" "$CODE"

echo
echo "=============================="
echo "Final Report"
echo "=============================="
echo "Passed: $PASS"
echo "Failed: $FAIL"

if [ "$FAIL" -eq 0 ]; then
  echo "✅ All tests passed"
  exit 0
else
  echo "❌ Some tests failed"
  echo "Last response body:"
  cat /tmp/bar108_response.json
  exit 1
fi
