#!/bin/bash

# Examples of using the Dogelytics API

BASE_URL="http://localhost:4420"

echo "=================================="
echo "Dogelytics API Examples"
echo "=================================="
echo ""

# Health Check
echo "1. Health Check:"
echo "   curl $BASE_URL/health"
curl -s "$BASE_URL/health" | jq '.'
echo ""
echo ""

# Balance Query Example
echo "2. Balance Query:"
echo "   Replace 'YOUR_DOGE_ADDRESS' with an actual Dogecoin address"
echo ""
echo "   Example with a mainnet address:"
echo "   curl '$BASE_URL/balance?address=D7Y55r1psU6xRfUyXRr59kAjdikKb8njf3'"
echo ""
echo "   Response format:"
echo "   {"
echo "     \"incoming\": \"100000000.00000000\",   // Pending balance (needs confirmations)"
echo "     \"available\": \"500000000.00000000\",  // Confirmed, spendable balance"
echo "     \"outgoing\": \"50000000.00000000\",    // Recently spent (needs confirmations)"
echo "     \"current\": \"600000000.00000000\"     // Total: incoming + available"
echo "   }"
echo ""

# Test with a known address (if provided as argument)
if [ ! -z "$1" ]; then
    echo "3. Testing with address: $1"
    curl -s "$BASE_URL/balance?address=$1" | jq '.'
    echo ""
fi

echo ""
echo "=================================="
echo "Usage: ./examples.sh [DOGE_ADDRESS]"
echo "=================================="

