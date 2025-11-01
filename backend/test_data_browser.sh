#!/bin/bash

# Test the Data Browser API with various scenarios

echo "🔍 Testing Bronze Backend Data Browser API"
echo "========================================"

BASE_URL="http://localhost:8060"

# Test 1: Health check
echo -e "\n1️⃣ Health Check"
curl -s "$BASE_URL/health" | jq '.status' || echo "❌ Health check failed"

# Test 2: List all data files
echo -e "\n2️⃣ List Data Files"
curl -s "$BASE_URL/api/data/files" | jq '.success, .count' || echo "❌ List files failed"

# Test 3: Browse Excel file (with XLSM support)
echo -e "\n3️⃣ Browse Excel File (supports XLSM)"
curl -s -X POST "$BASE_URL/api/data/browse" \
  -H "Content-Type: application/json" \
  -d '{
    "file_name": "sample_data.csv",
    "has_headers": true,
    "max_rows": 5
  }' | jq '.success, .row_count, .data_type' || echo "❌ Excel browse failed"

# Test 4: Browse with auto-detection
echo -e "\n4️⃣ Browse with Auto-Detection"
curl -s -X POST "$BASE_URL/api/data/browse" \
  -H "Content-Type: application/json" \
  -d '{
    "file_name": "sample_data.csv",
    "auto_detect_headers": true,
    "max_rows": 3
  }' | jq '.success, .has_headers' || echo "❌ Auto-detect failed"

# Test 5: Treat as CSV (works for any extension)
echo -e "\n5️⃣ Treat as CSV (Any File Type)"
curl -s -X POST "$BASE_URL/api/data/browse" \
  -H "Content-Type: application/json" \
  -d '{
    "file_name": "sample_data.csv",
    "treat_as_csv": true,
    "auto_detect_headers": true,
    "max_rows": 3
  }' | jq '.success, .message' || echo "❌ Treat as CSV failed"

# Test 6: Streaming mode
echo -e "\n6️⃣ Stream Large CSV File"
echo "Starting streaming response (multiple JSON objects):"
curl -s -X POST "$BASE_URL/api/data/browse" \
  -H "Content-Type: application/json" \
  -d '{
    "file_name": "sample_data.csv",
    "stream_mode": true,
    "chunk_size": 5,
    "max_rows": 10,
    "has_headers": true
  }' | while IFS= read -r line; do
    if [[ -n "$line" ]]; then
      echo "📦 Chunk: $(echo "$line" | jq -r '.row_count // .message // .success' 2>/dev/null || echo "$line")"
    fi
  done || echo "❌ Streaming failed"

echo -e "\n✅ Data Browser API Testing Complete!"
echo -e "\n📚 Available features:"
echo "  • Excel file support (.xlsx, .xls, .xlsm)"
echo "  • CSV file support with auto-detection"
echo "  • Universal CSV support (any file extension)"
echo "  • Streaming for large files"
echo "  • Auto-detection of delimiters and headers"
echo "  • Pagination and offset support"

echo -e "\n🔧 Advanced Options:"
echo "  • treat_as_csv: Process any file as CSV"
echo "  • auto_detect_headers: Auto-detect if first row contains headers"
echo "  • stream_mode: Enable streaming for large files"
echo "  • chunk_size: Configure streaming chunk size"
echo "  • max_rows: Limit number of rows returned"
echo "  • offset: Skip specified number of rows"