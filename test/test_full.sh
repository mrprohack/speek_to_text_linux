#!/bin/bash
# Full test suite for VoiceType

echo "╔════════════════════════════════════════════════════════════╗"
echo "║           VoiceType - Complete Test Suite                 ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""

PASS=0
FAIL=0

# Test 1: Check binary exists
echo "[Test 1/5] Checking VoiceType binaries..."
if [ -f ./VoiceType ] || [ -f ./VoiceType-gui ]; then
    echo "[PASS] VoiceType binary found"
    ((PASS++))
else
    echo "[FAIL] VoiceType binary not found"
    ((FAIL++))
fi

# Test 2: Check API key
echo ""
echo "[Test 2/5] Checking API key..."
if [ -n "$GROQ_API_KEY" ]; then
    echo "[PASS] GROQ_API_KEY is set"
    ((PASS++))
else
    echo "[WARN] GROQ_API_KEY not set"
    echo "       Set with: export GROQ_API_KEY='your_key'"
    ((FAIL++))
fi

# Test 3: Audio devices
echo ""
echo "[Test 3/5] Checking audio devices..."
if command -v arecord &> /dev/null; then
    DEVICES=$(arecord -l 2>/dev/null | grep -c "card" || echo "0")
    if [ "$DEVICES" -gt 0 ]; then
        echo "[PASS] Audio devices found: $DEVICES"
        ((PASS++))
    else
        echo "[WARN] No audio devices found"
        ((FAIL++))
    fi
else
    echo "[WARN] arecord not installed"
    ((FAIL++))
fi

# Test 4: Hotkey tool
echo ""
echo "[Test 4/5] Checking hotkey tools..."
if command -v xdotool &> /dev/null; then
    echo "[PASS] xdotool installed"
    ((PASS++))
else
    echo "[WARN] xdotool not installed"
    echo "       Install with: sudo apt install xdotool"
    ((FAIL++))
fi

# Test 5: Transcription test
echo ""
echo "[Test 5/5] Testing transcription API..."
if [ -n "$GROQ_API_KEY" ] && [ -f /tmp/voicetype_test.wav ]; then
    RESULT=$(curl -s -X POST \
        -H "Authorization: Bearer $GROQ_API_KEY" \
        -F "file=@/tmp/voicetype_test.wav" \
        -F "model=whisper-large-v3" \
        https://api.groq.com/openai/v1/audio/transcriptions)
    
    if echo "$RESULT" | grep -q "text"; then
        echo "[PASS] Transcription API works"
        ((PASS++))
    else
        echo "[FAIL] Transcription failed"
        ((FAIL++))
    fi
else
    echo "[SKIP] Skipped (missing API key or test file)"
    ((FAIL++))
fi

# Summary
echo ""
echo "╔════════════════════════════════════════════════════════════╗"
echo "║  Results: $PASS passed, $FAIL failed                      ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""

if [ $FAIL -eq 0 ]; then
    echo "All tests passed! VoiceType is ready to use."
    echo ""
    echo "Run with: ./VoiceType-gui"
else
    echo "Some tests failed. Check the output above for details."
    echo ""
    echo "Quick fixes:"
    echo "  - Set API key: export GROQ_API_KEY='your_key'"
    echo "  - Install xdotool: sudo apt install xdotool"
    echo "  - Create test audio: ./test/test_audio.sh"
fi
