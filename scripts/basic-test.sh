#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
SCREENSHOT_PATH="$PROJECT_DIR/test/doomerang-test-screenshot.png"

cd "$PROJECT_DIR"

# Ensure test directory exists
mkdir -p "$PROJECT_DIR/test"

echo "=== Doomerang Basic Test ==="
echo ""

# Step 1: Build the game
echo "[1/6] Building game..."
make build
echo "Build successful"
echo ""

# Step 2: Launch the game in background
echo "[2/6] Launching game..."
./doomerang &
GAME_PID=$!
echo "Game PID: $GAME_PID"
sleep 2

# Step 3: Verify game is running
echo "[3/6] Verifying game started..."
if ! ps -p $GAME_PID > /dev/null 2>&1; then
    echo "ERROR: Game crashed on startup"
    exit 1
fi
echo "Game is running"
echo ""

# Step 4: Focus game window and test movement
echo "[4/6] Testing input (start game, walk right + jump)..."
osascript <<'EOF'
tell application "System Events"
    set frontmost of process "doomerang" to true
    delay 0.5

    -- Press Enter to start game from main menu
    key code 36  -- return/enter key
    delay 1

    -- Walk right for 2 seconds using rapid key press loop
    set endTime to (current date) + 2
    repeat while (current date) < endTime
        key code 124  -- right arrow
    end repeat

    delay 0.2

    -- Jump once
    key code 7  -- x key
end tell
EOF
echo "Input test completed"
echo ""

# Step 5: Take screenshot for verification
echo "[5/6] Taking screenshot..."
# Get window bounds of the frontmost window (doomerang) and capture that region
BOUNDS=$(osascript -e '
tell application "System Events"
    set frontApp to first process whose frontmost is true
    set winPos to position of window 1 of frontApp
    set winSize to size of window 1 of frontApp
    set x to item 1 of winPos
    set y to item 2 of winPos
    set w to item 1 of winSize
    set h to item 2 of winSize
    return (x as text) & "," & (y as text) & "," & (w as text) & "," & (h as text)
end tell
' 2>/dev/null)

if [ -n "$BOUNDS" ]; then
    screencapture -R "$BOUNDS" "$SCREENSHOT_PATH"
else
    echo "Warning: Could not get window bounds, taking full screen capture"
    screencapture -x "$SCREENSHOT_PATH"
fi
echo "Screenshot saved to: $SCREENSHOT_PATH"
echo ""

# Step 6: Verify game still running (no crash during test)
echo "[6/6] Verifying no crash during test..."
if ! ps -p $GAME_PID > /dev/null 2>&1; then
    echo "ERROR: Game crashed during testing"
    exit 1
fi
echo "Game completed test without crashing"
echo ""

# Cleanup
echo "Cleaning up..."
kill $GAME_PID 2>/dev/null || true
echo ""

echo "=== SUCCESS: All basic tests passed ==="
echo "Screenshot available at: $SCREENSHOT_PATH"
