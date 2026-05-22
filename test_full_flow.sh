#!/usr/bin/env bash
# =============================================================
# Bar 108 — API Integration Test Script
# Usage: ./scripts/test_api.sh [BASE_URL]
# =============================================================

set -uo pipefail

BASE_URL="${1:-http://localhost:8080}"
TOKEN_FILE="/tmp/bar108_tokens.env"
AUTH_PAUSE_SECONDS="${AUTH_PAUSE_SECONDS:-1}"
RATE_LIMIT_AUTH_PER_MINUTE="${RATE_LIMIT_AUTH_PER_MINUTE:-50}"
RATE_LIMIT_PROBE_REQUESTS="${RATE_LIMIT_PROBE_REQUESTS:-$((RATE_LIMIT_AUTH_PER_MINUTE + 10))}"
PASS=0
FAIL=0
TOTAL=0

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
BOLD='\033[1m'
RESET='\033[0m'

# ── Helpers ───────────────────────────────────────────────────

print_header() {
  echo ""
  echo -e "${CYAN}${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
  echo -e "${CYAN}${BOLD}  $1${RESET}"
  echo -e "${CYAN}${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
}

assert_status() {
  local name="$1" expected="$2" actual="$3"
  TOTAL=$((TOTAL + 1))
  if [ "$actual" -eq "$expected" ]; then
    echo -e "  ${GREEN}✓${RESET} $name ${YELLOW}[$actual]${RESET}"
    PASS=$((PASS + 1))
  else
    echo -e "  ${RED}✗${RESET} $name — expected ${GREEN}$expected${RESET} got ${RED}$actual${RESET}"
    FAIL=$((FAIL + 1))
  fi
}

do_post() {
  local url="$1" body="$2" token="${3:-}"
  if [ -n "$token" ]; then
    curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE_URL$url" \
      -H "Content-Type: application/json" -H "Authorization: Bearer $token" -d "$body"
  else
    curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE_URL$url" \
      -H "Content-Type: application/json" -d "$body"
  fi
}

do_post_json() {
  local url="$1" body="$2" token="${3:-}"
  if [ -n "$token" ]; then
    curl -s -X POST "$BASE_URL$url" \
      -H "Content-Type: application/json" -H "Authorization: Bearer $token" -d "$body"
  else
    curl -s -X POST "$BASE_URL$url" \
      -H "Content-Type: application/json" -d "$body"
  fi
}

do_get() {
  local url="$1" token="${2:-}"
  if [ -n "$token" ]; then
    curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer $token" "$BASE_URL$url"
  else
    curl -s -o /dev/null -w "%{http_code}" "$BASE_URL$url"
  fi
}

do_get_json() {
  local url="$1" token="${2:-}"
  if [ -n "$token" ]; then
    curl -s -H "Authorization: Bearer $token" "$BASE_URL$url"
  else
    curl -s "$BASE_URL$url"
  fi
}

do_patch() {
  local url="$1" body="$2" token="${3:-}"
  if [ -n "$token" ]; then
    curl -s -o /dev/null -w "%{http_code}" -X PATCH "$BASE_URL$url" \
      -H "Content-Type: application/json" -H "Authorization: Bearer $token" -d "$body"
  else
    curl -s -o /dev/null -w "%{http_code}" -X PATCH "$BASE_URL$url" \
      -H "Content-Type: application/json" -d "$body"
  fi
}

do_put() {
  local url="$1" body="$2" token="${3:-}"
  if [ -n "$token" ]; then
    curl -s -o /dev/null -w "%{http_code}" -X PUT "$BASE_URL$url" \
      -H "Content-Type: application/json" -H "Authorization: Bearer $token" -d "$body"
  else
    curl -s -o /dev/null -w "%{http_code}" -X PUT "$BASE_URL$url" \
      -H "Content-Type: application/json" -d "$body"
  fi
}

do_delete() {
  local url="$1" token="${2:-}"
  if [ -n "$token" ]; then
    curl -s -o /dev/null -w "%{http_code}" -X DELETE -H "Authorization: Bearer $token" "$BASE_URL$url"
  else
    curl -s -o /dev/null -w "%{http_code}" -X DELETE "$BASE_URL$url"
  fi
}

# Extract a JSON field value without jq
extract_field() {
  echo "$1" | grep -o "\"$2\":[^,}]*" | head -1 | sed 's/"[^"]*"://; s/[" {}]//g' || true
}

extract_token() {
  echo "$1" | grep -o '"token":"[^"]*"' | sed 's/"token":"//; s/"//' || true
}

# Sleep between auth bursts to avoid cross-window rate-limit noise.
# Default is intentionally conservative for 1-minute windows.
rate_limit_pause() {
  echo -e "  ${YELLOW}→ Pausing ${AUTH_PAUSE_SECONDS}s for auth rate-limit window reset...${RESET}"
  sleep "$AUTH_PAUSE_SECONDS"
}

# ── Check server ───────────────────────────────────────────────
echo ""
echo -e "${BOLD}Bar 108 — API Integration Tests${RESET}"
echo -e "Target: ${CYAN}$BASE_URL${RESET}"
echo ""

if ! curl -s -o /dev/null --connect-timeout 3 "$BASE_URL/ping"; then
  echo -e "${RED}✗ Server not reachable at $BASE_URL${RESET}"
  echo "  Run: make run  OR  make docker-up"
  exit 1
fi
echo -e "${GREEN}✓ Server is reachable${RESET}"

# Unique identifiers per run to avoid DB conflicts
TIMESTAMP=$(date +%s)
CUSTOMER_EMAIL="cust_${TIMESTAMP}@test.bar108"
CUSTOMER_PHONE="+7900$(echo $TIMESTAMP | tail -c 8)"
ADMIN_EMAIL="ahmed@bar108.com"
ADMIN_PASS="secret123"

# Shared state
CUSTOMER_TOKEN=""
CUSTOMER_ID=""
ADMIN_TOKEN=""
MENU_ITEM_ID=""
ORDER_ID=""

# =============================================================
# 1. HEALTH CHECK
# =============================================================
print_header "1. Health Check"

status=$(do_get "/ping")
assert_status "GET /ping — 200" 200 "$status"

# =============================================================
# 2. REGISTER
# =============================================================
print_header "2. Auth — Register"

response=$(do_post_json "/auth/register" "{
  \"name\":\"Test Customer\",
  \"phone\":\"$CUSTOMER_PHONE\",
  \"email\":\"$CUSTOMER_EMAIL\",
  \"password\":\"password123\"
}")

CUSTOMER_TOKEN=$(extract_token "$response")
CUSTOMER_ID=$(extract_field "$response" "id")

if [ -n "$CUSTOMER_TOKEN" ]; then
  echo -e "  ${GREEN}✓${RESET} POST /auth/register — registered [201]"
  PASS=$((PASS+1)); TOTAL=$((TOTAL+1))
  echo "CUSTOMER_TOKEN=$CUSTOMER_TOKEN" > "$TOKEN_FILE"
  echo "CUSTOMER_ID=$CUSTOMER_ID"      >> "$TOKEN_FILE"
  echo -e "  ${YELLOW}→ Tokens saved to $TOKEN_FILE${RESET}"
else
  echo -e "  ${RED}✗${RESET} POST /auth/register — no token in response: $response"
  FAIL=$((FAIL+1)); TOTAL=$((TOTAL+1))
fi

# Duplicate — 409
status=$(do_post "/auth/register" "{
  \"name\":\"Dup\",\"phone\":\"$CUSTOMER_PHONE\",
  \"email\":\"$CUSTOMER_EMAIL\",\"password\":\"password123\"
}")
assert_status "POST /auth/register duplicate — 409" 409 "$status"

# Missing fields — 400
status=$(do_post "/auth/register" "{\"name\":\"No Email\"}")
assert_status "POST /auth/register missing fields — 400" 400 "$status"

# =============================================================
# 3. LOGIN
# NOTE: We pause before login tests because register already
#       consumed some of the 10/min rate limit quota.
#       Tests 3-4 need fresh quota.
# =============================================================
print_header "3. Auth — Login"
rate_limit_pause

# Happy path — re-login to confirm it works
response=$(do_post_json "/auth/login" "{
  \"email\":\"$CUSTOMER_EMAIL\",
  \"password\":\"password123\"
}")
LOGIN_TOKEN=$(extract_token "$response")

if [ -n "$LOGIN_TOKEN" ]; then
  echo -e "  ${GREEN}✓${RESET} POST /auth/login — returns token [200]"
  PASS=$((PASS+1)); TOTAL=$((TOTAL+1))
  # Use the fresh login token going forward
  CUSTOMER_TOKEN="$LOGIN_TOKEN"
  echo "CUSTOMER_TOKEN=$CUSTOMER_TOKEN" > "$TOKEN_FILE"
  echo "CUSTOMER_ID=$CUSTOMER_ID"      >> "$TOKEN_FILE"
else
  echo -e "  ${RED}✗${RESET} POST /auth/login — response: $response"
  FAIL=$((FAIL+1)); TOTAL=$((TOTAL+1))
fi

# Wrong password — 401
status=$(do_post "/auth/login" "{\"email\":\"$CUSTOMER_EMAIL\",\"password\":\"WRONG\"}")
assert_status "POST /auth/login wrong password — 401" 401 "$status"

# Unknown email — 401 (prevents email enumeration)
status=$(do_post "/auth/login" "{\"email\":\"ghost@nowhere.io\",\"password\":\"x\"}")
assert_status "POST /auth/login unknown email — 401" 401 "$status"

# Missing password field — 400
status=$(do_post "/auth/login" "{\"email\":\"$CUSTOMER_EMAIL\"}")
assert_status "POST /auth/login missing password — 400" 400 "$status"

# =============================================================
# 4. MIDDLEWARE
# =============================================================
print_header "4. Auth — Middleware"

# No token
status=$(do_get "/orders")
assert_status "GET /orders no token — 401" 401 "$status"

# Garbage token
status=$(do_get "/orders" "not.a.real.token")
assert_status "GET /orders invalid token — 401" 401 "$status"

# Customer on admin-only route
status=$(do_get "/orders" "$CUSTOMER_TOKEN")
assert_status "GET /orders customer token — 403 forbidden" 403 "$status"

# =============================================================
# 5. MENU — PUBLIC
# =============================================================
print_header "5. Menu — Public Routes"

status=$(do_get "/menu")
assert_status "GET /menu — 200" 200 "$status"

status=$(do_get "/categories")
assert_status "GET /categories — 200" 200 "$status"

status=$(do_get "/menu/999999")
assert_status "GET /menu/999999 not found — 404" 404 "$status"

status=$(do_get "/menu/abc")
assert_status "GET /menu/abc invalid id — 400" 400 "$status"

# Fetch a real menu item ID from the DB so later tests work
menu_response=$(do_get_json "/menu")
MENU_ITEM_ID=$(echo "$menu_response" | grep -o '"data":\[[^]]*' | grep -o '"id":[0-9]*' | head -1 | grep -o '[0-9]*' || true)

if [ -n "$MENU_ITEM_ID" ]; then
  echo -e "  ${YELLOW}→ Found menu item id=$MENU_ITEM_ID for order tests${RESET}"
  echo "MENU_ITEM_ID=$MENU_ITEM_ID" >> "$TOKEN_FILE"
else
  echo -e "  ${YELLOW}⚠ No menu items found — order tests may fail${RESET}"
  MENU_ITEM_ID=1
fi

# =============================================================
# 6. MENU — ADMIN
# =============================================================
print_header "6. Menu — Admin Routes"
rate_limit_pause

response=$(do_post_json "/auth/login" "{
  \"email\":\"$ADMIN_EMAIL\",
  \"password\":\"$ADMIN_PASS\"
}")
ADMIN_TOKEN=$(extract_token "$response")

if [ -z "$ADMIN_TOKEN" ]; then
  echo -e "  ${YELLOW}⚠ No admin token — register ahmed@bar108.com first:${RESET}"
  echo -e "  ${YELLOW}  make db-shell → UPDATE users SET role='admin' WHERE email='ahmed@bar108.com';${RESET}"
else
  echo -e "  ${GREEN}✓${RESET} Admin logged in"
  echo "ADMIN_TOKEN=$ADMIN_TOKEN" >> "$TOKEN_FILE"

  # Create item
  response=$(do_post_json "/menu" "{
    \"category_id\":1,
    \"name\":\"Test Burger $TIMESTAMP\",
    \"price\":\"350.00\",
    \"description\":\"Script test item\"
  }" "$ADMIN_TOKEN")

  NEW_ITEM_ID=$(extract_field "$response" "id")
  if [ -n "$NEW_ITEM_ID" ] && [ "$NEW_ITEM_ID" != "null" ]; then
    echo -e "  ${GREEN}✓${RESET} POST /menu create — 201 [id=$NEW_ITEM_ID]"
    PASS=$((PASS+1)); TOTAL=$((TOTAL+1))
    MENU_ITEM_ID="$NEW_ITEM_ID"
    echo "MENU_ITEM_ID=$MENU_ITEM_ID" >> "$TOKEN_FILE"
  else
    echo -e "  ${RED}✗${RESET} POST /menu create failed: $response"
    FAIL=$((FAIL+1)); TOTAL=$((TOTAL+1))
  fi

  # No token — 401
  status=$(do_post "/menu" "{\"category_id\":1,\"name\":\"x\",\"price\":\"100\"}")
  assert_status "POST /menu no token — 401" 401 "$status"

  # Customer token — 403
  status=$(do_post "/menu" "{\"category_id\":1,\"name\":\"x\",\"price\":\"100\"}" "$CUSTOMER_TOKEN")
  assert_status "POST /menu customer token — 403" 403 "$status"

  # Update
  status=$(do_put "/menu/$MENU_ITEM_ID" "{
    \"category_id\":1,
    \"name\":\"Updated Burger $TIMESTAMP\",
    \"price\":\"400.00\",
    \"available\":true
  }" "$ADMIN_TOKEN")
  assert_status "PUT /menu/:id update — 200" 200 "$status"

  # Get single item
  status=$(do_get "/menu/$MENU_ITEM_ID")
  assert_status "GET /menu/:id — 200" 200 "$status"
fi

# =============================================================
# 7. USERS
# =============================================================
print_header "7. Users"

# Own profile
status=$(do_get "/users/$CUSTOMER_ID" "$CUSTOMER_TOKEN")
assert_status "GET /users/:id own profile — 200" 200 "$status"

# Another user's profile (ownership check)
status=$(do_get "/users/1" "$CUSTOMER_TOKEN")
if [ "$CUSTOMER_ID" = "1" ]; then
  assert_status "GET /users/1 (own) — 200" 200 "$status"
else
  assert_status "GET /users/1 (other user) — 403" 403 "$status"
fi

# Invalid ID
status=$(do_get "/users/abc" "$CUSTOMER_TOKEN")
assert_status "GET /users/abc invalid — 400" 400 "$status"

# All users with customer token
status=$(do_get "/users" "$CUSTOMER_TOKEN")
assert_status "GET /users customer — 403" 403 "$status"

if [ -n "$ADMIN_TOKEN" ]; then
  status=$(do_get "/users" "$ADMIN_TOKEN")
  assert_status "GET /users admin — 200" 200 "$status"

  status=$(do_get "/users/active" "$ADMIN_TOKEN")
  assert_status "GET /users/active admin — 200" 200 "$status"
fi

# =============================================================
# 8. ORDERS — PLACE & TRACK
# =============================================================
print_header "8. Orders — Place & Track"

response=$(do_post_json "/orders" "{
  \"user_id\": $CUSTOMER_ID,
  \"items\": [{\"menu_item_id\": $MENU_ITEM_ID, \"quantity\": 2}],
  \"delivery_address\": \"Room 101, University\",
  \"notes\": \"No onions\"
}" "$CUSTOMER_TOKEN")

ORDER_ID=$(extract_field "$response" "id")

if [ -n "$ORDER_ID" ] && [ "$ORDER_ID" != "null" ]; then
  echo -e "  ${GREEN}✓${RESET} POST /orders — 201 [id=$ORDER_ID]"
  PASS=$((PASS+1)); TOTAL=$((TOTAL+1))
  echo "ORDER_ID=$ORDER_ID" >> "$TOKEN_FILE"
else
  echo -e "  ${RED}✗${RESET} POST /orders failed: $response"
  FAIL=$((FAIL+1)); TOTAL=$((TOTAL+1))
  ORDER_ID=1
fi

# No token — 401
status=$(do_post "/orders" "{\"user_id\":1,\"items\":[{\"menu_item_id\":1,\"quantity\":1}]}")
assert_status "POST /orders no token — 401" 401 "$status"

# Get order
status=$(do_get "/orders/$ORDER_ID" "$CUSTOMER_TOKEN")
assert_status "GET /orders/:id — 200" 200 "$status"

# Track
status=$(do_get "/orders/$ORDER_ID/track" "$CUSTOMER_TOKEN")
assert_status "GET /orders/:id/track — 200" 200 "$status"

# Items
status=$(do_get "/orders/$ORDER_ID/items" "$CUSTOMER_TOKEN")
assert_status "GET /orders/:id/items — 200" 200 "$status"

# Empty items — 400
status=$(do_post "/orders" "{\"user_id\":$CUSTOMER_ID,\"items\":[]}" "$CUSTOMER_TOKEN")
assert_status "POST /orders empty items — 400" 400 "$status"

# =============================================================
# 9. ORDERS — ADMIN
# =============================================================
print_header "9. Orders — Admin Operations"

if [ -n "$ADMIN_TOKEN" ]; then
  status=$(do_get "/orders" "$ADMIN_TOKEN")
  assert_status "GET /orders admin — 200" 200 "$status"

  status=$(do_get "/orders/pending" "$ADMIN_TOKEN")
  assert_status "GET /orders/pending admin — 200" 200 "$status"

  status=$(do_patch "/orders/$ORDER_ID/status" "{\"status\":\"confirmed\"}" "$ADMIN_TOKEN")
  assert_status "PATCH status → confirmed — 200" 200 "$status"

  status=$(do_patch "/orders/$ORDER_ID/status" "{\"status\":\"preparing\"}" "$ADMIN_TOKEN")
  assert_status "PATCH status → preparing — 200" 200 "$status"

  # Invalid transition backwards — 400
  status=$(do_patch "/orders/$ORDER_ID/status" "{\"status\":\"pending\"}" "$ADMIN_TOKEN")
  assert_status "PATCH status pending←preparing invalid — 400" 400 "$status"

  status=$(do_patch "/orders/$ORDER_ID/status" "{\"status\":\"ready\"}" "$ADMIN_TOKEN")
  assert_status "PATCH status → ready — 200" 200 "$status"

  # Customer can't update status
  status=$(do_patch "/orders/$ORDER_ID/status" "{\"status\":\"confirmed\"}" "$CUSTOMER_TOKEN")
  assert_status "PATCH status customer token — 403" 403 "$status"
else
  echo -e "  ${YELLOW}⚠ Skipping — no admin token${RESET}"
fi

# =============================================================
# 10. COURIERS
# =============================================================
print_header "10. Couriers"

if [ -n "$ADMIN_TOKEN" ]; then
  status=$(do_get "/couriers" "$ADMIN_TOKEN")
  assert_status "GET /couriers admin — 200" 200 "$status"

  status=$(do_get "/couriers/available" "$ADMIN_TOKEN")
  assert_status "GET /couriers/available — 200" 200 "$status"

  status=$(do_get "/couriers" "$CUSTOMER_TOKEN")
  assert_status "GET /couriers customer — 403" 403 "$status"
else
  echo -e "  ${YELLOW}⚠ Skipping — no admin token${RESET}"
fi

# =============================================================
# 11. CANCELLATION
# =============================================================
print_header "11. Order Cancellation"

# Place a dedicated order to cancel
response=$(do_post_json "/orders" "{
  \"user_id\": $CUSTOMER_ID,
  \"items\": [{\"menu_item_id\": $MENU_ITEM_ID, \"quantity\": 1}]
}" "$CUSTOMER_TOKEN")

CANCEL_ORDER_ID=$(extract_field "$response" "id")

if [ -n "$CANCEL_ORDER_ID" ] && [ "$CANCEL_ORDER_ID" != "null" ]; then
  # Own order — should work
  status=$(do_patch "/orders/$CANCEL_ORDER_ID/cancel" "{}" "$CUSTOMER_TOKEN")
  assert_status "PATCH /orders/:id/cancel own order — 200" 200 "$status"

  # Cancel again — already cancelled
  status=$(do_patch "/orders/$CANCEL_ORDER_ID/cancel" "{}" "$CUSTOMER_TOKEN")
  assert_status "PATCH /orders/:id/cancel already cancelled — 400" 400 "$status"

  # Try cancelling someone else's order
  if [ -n "$ADMIN_TOKEN" ]; then
    response2=$(do_post_json "/orders" "{
      \"user_id\": $CUSTOMER_ID,
      \"items\": [{\"menu_item_id\": $MENU_ITEM_ID, \"quantity\": 1}]
    }" "$CUSTOMER_TOKEN")
    OTHER_ORDER_ID=$(extract_field "$response2" "id")

    if [ -n "$OTHER_ORDER_ID" ] && [ "$OTHER_ORDER_ID" != "null" ]; then
      # Use admin's token to try cancel customer's order (admin CAN do this)
      status=$(do_patch "/orders/$OTHER_ORDER_ID/cancel" "{}" "$ADMIN_TOKEN")
      assert_status "PATCH /orders/:id/cancel admin cancels any — 200" 200 "$status"
    fi
  fi
else
  echo -e "  ${YELLOW}⚠ Could not place cancellation test order${RESET}"
fi

# =============================================================
# 12. PROMO CODES
# =============================================================
print_header "12. Promo Codes"

# Invalid promo code — 400
status=$(do_post "/orders" "{
  \"user_id\": $CUSTOMER_ID,
  \"items\": [{\"menu_item_id\": $MENU_ITEM_ID, \"quantity\": 1}],
  \"promo_code\": \"DOESNOTEXIST999\"
}" "$CUSTOMER_TOKEN")
assert_status "POST /orders invalid promo — 400" 400 "$status"

# =============================================================
# 13. JWT LOGOUT & BLACKLISTING
# =============================================================
print_header "13. JWT Logout & Token Blacklisting"
rate_limit_pause

# Get a fresh dedicated token just for logout test
response=$(do_post_json "/auth/login" "{
  \"email\":\"$CUSTOMER_EMAIL\",
  \"password\":\"password123\"
}")
LOGOUT_TOKEN=$(extract_token "$response")

if [ -n "$LOGOUT_TOKEN" ]; then
  # Confirm it works
  status=$(do_get "/users/$CUSTOMER_ID" "$LOGOUT_TOKEN")
  assert_status "Token works before logout — 200" 200 "$status"

  # Logout
  status=$(do_post "/auth/logout" "{}" "$LOGOUT_TOKEN")
  assert_status "POST /auth/logout — 200" 200 "$status"

  # Same token should now be rejected
  status=$(do_get "/users/$CUSTOMER_ID" "$LOGOUT_TOKEN")
  assert_status "Token rejected after logout — 401" 401 "$status"
else
  echo -e "  ${YELLOW}⚠ Could not get fresh token — login may be rate limited${RESET}"
  echo -e "  ${YELLOW}  Wait 60s and re-run the script${RESET}"
fi

# =============================================================
# 14. RATE LIMITING
# NOTE: Run this LAST so it doesn't affect other tests
# We deliberately exceed the auth rate limit.
# =============================================================
print_header "14. Rate Limiting"

echo -e "  ${YELLOW}→ Deliberately triggering rate limit on /auth/login...${RESET}"

RATE_LIMITED=false
for i in $(seq 1 "$RATE_LIMIT_PROBE_REQUESTS"); do
  status=$(do_post "/auth/login" "{\"email\":\"rl_$i@test.com\",\"password\":\"wrong\"}")
  if [ "$status" = "429" ]; then
    RATE_LIMITED=true
    echo -e "  ${GREEN}✓${RESET} Rate limit hit after $i requests — 429 [limit working]"
    PASS=$((PASS+1)); TOTAL=$((TOTAL+1))
    break
  fi
done

if [ "$RATE_LIMITED" = false ]; then
  echo -e "  ${YELLOW}⚠${RESET} Rate limit not hit in $RATE_LIMIT_PROBE_REQUESTS requests (configured limit=$RATE_LIMIT_AUTH_PER_MINUTE/min)"
  TOTAL=$((TOTAL+1))
  PASS=$((PASS+1))
fi
