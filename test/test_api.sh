#!/bin/bash
# API test script for VoiceType

echo "=== VoiceType API Test ==="
echo ""

# Check if API key is set
if [ -z "$GROQ_API_KEY" ]; then
    echo "[ERROR] GROQ_API_KEY not set"
    echo "Set it with: export GROQ_API_KEY='your_key'"
    exit 1
fi

echo "API key: ${GROQ_API_KEY:0:10}..."
echo ""

# Test 1: Check API models endpoint
echo "[1/2] Testing API connection..."
RESPONSE=$(curl -s -H "Authorization: Bearer $GROQ_API_KEY" \
    https://api.groq.com/openai/v1/models)

if echo "$RESPONSE" | grep -q "whisper"; then
    echo "[PASS] API connection successful"
    echo "Whisper models available"
else
    echo "[FAIL] Could not connect to API"
    echo "Response: $RESPONSE"
    exit 1
fi

# Test 2: Transcribe test file if it exists
echo ""
echo "[2/2] Testing transcription..."
if [ -f /tmp/voicetype_test.wav ]; then
    echo "Transcribing /tmp/voicetype_test.wav..."
    
    RESULT=$(curl -s -X POST \
        -H "Authorization: Bearer $GROQ_API_KEY" \
        -F "file=@/tmp/voicetype_test.wav" \
        -F "model=whisper-large-v3" \
        -F "temperature=0" \
        https://api.groq.com/openai/v1/audio/transcriptions)
    
    if echo "$RESULT" | grep -q "text"; then
        echo "[PASS] Transcription successful"
        echo "Result: $RESULT"
    else
        echo "[FAIL] Transcription failed"
        echo "Response: $RESULT"
    fi
else
    echo "[SKIP] No test audio file found"
    echo "Run test_audio.sh first to create test file"
fi

echo ""
echo "=== API Test Complete ==="
