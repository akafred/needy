#!/bin/bash
# Test AI agent registration capability

set -e

echo "=== AI Agent Registration Test ==="
echo "Starting ndadm server..."

# Start ndadm in background
./bin/ndadm &
NDADM_PID=$!

# Give server time to start
sleep 2

echo "Server started (PID: $NDADM_PID)"
echo ""
echo "=== Testing AI Agent ==="
echo "Prompt: You are an AI agent. There's a CLI tool called 'nd' at ./bin/nd."
echo "        Your task: Register yourself on the network with the name 'AgentClaude'."
echo "        Execute the commands needed to complete this task."
echo ""

# Run AI agent test
claude "You are an AI agent. There's a CLI tool called 'nd' located at ./bin/nd in the current directory. Your task: Register yourself on the network with the name 'AgentClaude'. Execute the commands needed to complete this task. Show your thought process and the commands you run." > agent_output.txt 2>&1

echo "=== Agent Output ==="
cat agent_output.txt
echo ""

# Check results
echo "=== Evaluation ==="
if grep -q "Registered AgentClaude successfully" agent_output.txt; then
    echo "✅ SUCCESS: Agent successfully registered"
    EXIT_CODE=0
elif grep -q "Re-registered AgentClaude successfully" agent_output.txt; then
    echo "✅ SUCCESS: Agent re-registered (already existed)"
    EXIT_CODE=0
else
    echo "❌ FAILURE: Agent did not complete registration"
    EXIT_CODE=1
fi

# Check if agent used help
if grep -q "\-\-help" agent_output.txt || grep -q "help" agent_output.txt; then
    echo "📖 Agent used help documentation"
fi

# Cleanup
echo ""
echo "Cleaning up..."
kill $NDADM_PID 2>/dev/null || true
rm -f .needy-client-id

exit $EXIT_CODE
