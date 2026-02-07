# AI Agent Testing Framework

## Goal
Test if AI models can understand and use the Needy CLI with minimal prompting, validating the UX and discoverability.

## Approach

### Test Operator (Human)
- Runs `ndadm` server
- Monitors agent behavior
- Evaluates success/failure

### Test Agents (AI Models via `claude`)
- Receive minimal context about their role
- Must discover and use CLI commands
- Evaluated on ability to complete tasks

## Test Scenarios

### Level 1: Registration (Available Now)
**Agent Prompt:**
```
You are an AI agent. There's a CLI tool called `nd` in your PATH. 
Your task: Register yourself on the network with the name "Agent[YourName]".
```

**Success Criteria:**
- Agent discovers `nd` help
- Agent runs `nd register --name Agent[Name]`
- Agent successfully registers

**Evaluation:**
- Did agent use `--help`?
- How many attempts to register?
- Did agent understand error messages?

### Level 2: Basic Communication (After Implementation)
**Agent Prompt:**
```
You are Agent Alice. You've been registered on the Needy network.
Your task: Broadcast a need for help with "fixing authentication bug".
```

**Success Criteria:**
- Agent discovers `nd send` command
- Agent sends need with correct syntax
- Message is broadcast successfully

### Level 3: Full Workflow (After Implementation)
**Setup:**
- Operator sends a need as AgentAlice
- AI agent is AgentBob

**Agent Prompt:**
```
You are Agent Bob on the Needy network. 
Check for any needs from other agents and help if you can.
```

**Success Criteria:**
- Agent discovers `nd receive`
- Agent sees the need
- Agent announces intent
- Agent provides solution
- Complete workflow executed

## Test Script Structure

```bash
#!/bin/bash
# test-ai-agent.sh

# Start ndadm
./bin/ndadm &
NDADM_PID=$!
sleep 1

# Test Level 1: Registration
echo "=== Testing AI Agent Registration ==="
claude "You are an AI agent. There's a CLI tool called 'nd' in your PATH at ./bin/nd. Your task: Register yourself on the network with the name 'AgentClaude'. Execute the commands needed." > agent_output.txt

# Check if registration succeeded
if grep -q "Registered AgentClaude successfully" agent_output.txt; then
    echo "✅ Agent successfully registered"
else
    echo "❌ Agent failed to register"
fi

# Cleanup
kill $NDADM_PID
```

## Models to Test

1. **Claude 3.5 Sonnet** (baseline)
2. **Claude 3 Opus** (comparison)
3. **GPT-4** (if available)
4. **Smaller models** (to test minimum capability)

## Metrics

- **Discovery Rate**: % of agents that find the right commands
- **Attempts to Success**: Number of tries before success
- **Error Recovery**: Can agents recover from mistakes?
- **Help Usage**: Do agents use `--help` flags?
- **Workflow Completion**: % completing full need→intent→solution

## Implementation Plan

1. ✅ **Plan framework** (this document)
2. **Test registration** - Create script for Level 1
3. **Finish communication** - Complete `nd send`/`receive`
4. **Test communication** - Create scripts for Level 2 & 3
5. **Analyze results** - Document which models succeed

## Expected Insights

- Which command patterns are intuitive?
- Where do agents get stuck?
- What error messages are confusing?
- Is `--help` output sufficient?
- Are command names discoverable?

This will drive UX improvements based on actual AI agent behavior.
