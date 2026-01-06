#!/bin/bash
# Audio test script for VoiceType

echo "=== VoiceType Audio Test ==="
echo ""

# Check if arecord is available
if ! command -v arecord &> /dev/null; then
    echo "[ERROR] arecord not found. Install with: sudo apt install alsa-utils"
    exit 1
fi

# Test 1: List audio devices
echo "[1/3] Checking audio devices..."
echo "Available audio devices:"
arecord -l
echo ""

# Test 2: Record a short sample
echo "[2/3] Recording test (2 seconds)..."
arecord -d 2 -f cd -t wav -r 16000 /tmp/voicetype_test.wav 2>/dev/null

if [ -f /tmp/voicetype_test.wav ]; then
    echo "Test recording saved to /tmp/voicetype_test.wav"
    SIZE=$(stat -c%s /tmp/voicetype_test.wav)
    echo "File size: $SIZE bytes"
    echo "[PASS] Audio recording works"
else
    echo "[FAIL] Could not record audio"
    exit 1
fi

echo ""
echo "[3/3] Audio test complete!"
echo ""
echo "To test with VoiceType:"
echo "  ./VoiceType-gui"
echo "  Press Ctrl+Space to record, release to transcribe"
