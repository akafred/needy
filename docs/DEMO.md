# Needy CLI Demo Tutorial

This tutorial demonstrates the agent registration system with hands-on examples.

## Prerequisites

Build the project:
```bash
make build
```

## Demo 1: Single Agent Registration

### Step 1: Start the ndadm server

In terminal 1:
```bash
./bin/ndadm
```

You should see:
```
ndadm: NATS server started on port 4222
ndadm: Listening for agent registrations...
```

### Step 2: Register an agent

In terminal 2:
```bash
./bin/nd register --name AgentAlice
```

Expected output:
```
Registered AgentAlice successfully

Now you can use the following commands to communicate:
  send [type] [message]  Broadcast a message (need, intent, or solution)
  receive                Read your unread messages
```

The server terminal will show:
```
ndadm: Registered agent 'AgentAlice' with client ID <uuid>
```

### Step 3: Verify client ID persistence

Check that a client ID file was created:
```bash
cat .needy-client-id
```

You'll see a UUID like: `550e8400-e29b-41d4-a716-446655440000`

### Step 4: Re-register (same client)

Run the same command again:
```bash
./bin/nd register --name AgentAlice
```

Expected output:
```
Re-registered AgentAlice successfully
```

Notice it says "Re-registered" instead of "Registered" - the system recognized your client ID!

## Demo 2: Multiple Agents

### Step 1: Create separate directories for each agent

```bash
mkdir -p /tmp/demo-alice /tmp/demo-bob /tmp/demo-charlie
```

### Step 2: Register multiple agents

In terminal 2 (Alice):
```bash
cd /tmp/demo-alice
/path/to/needy/bin/nd register --name AgentAlice
```

In terminal 3 (Bob):
```bash
cd /tmp/demo-bob
/path/to/needy/bin/nd register --name AgentBob
```

In terminal 4 (Charlie):
```bash
cd /tmp/demo-charlie
/path/to/needy/bin/nd register --name AgentCharlie
```

The server will show all three registrations:
```
ndadm: Registered agent 'AgentAlice' with client ID <uuid-1>
ndadm: Registered agent 'AgentBob' with client ID <uuid-2>
ndadm: Registered agent 'AgentCharlie' with client ID <uuid-3>
```

## Demo 3: Impersonation Prevention

### Step 1: Register an agent

```bash
cd /tmp/demo-alice
./bin/nd register --name AgentAlice
```

### Step 2: Try to impersonate from a different location

```bash
cd /tmp/demo-bob
./bin/nd register --name AgentAlice
```

Expected output:
```
Error: Agent name 'AgentAlice' is already registered
```

The server protects against impersonation by validating client IDs!

## Demo 4: Network Failure Handling

### Step 1: Stop the ndadm server

Press `Ctrl+C` in the ndadm terminal.

### Step 2: Try to register

```bash
./bin/nd register --name AgentDave
```

Expected output:
```
Error: Could not connect to network
(Details: nats: no servers available for connection)
```

The client handles network failures gracefully.

## Demo 5: Error Handling

### Step 1: Try to register without a name

```bash
./bin/nd register
```

Expected output:
```
Error: --name flag is required
Usage: nd register --name [name]
```

## Cleanup

Stop the ndadm server with `Ctrl+C` and clean up demo directories:
```bash
rm -rf /tmp/demo-alice /tmp/demo-bob /tmp/demo-charlie
```

## What's Next?

The Registration Phase is complete! Future phases will add:
- **Communication Phase**: Send and receive messages (needs, intents, solutions)
- **Payload Management**: Attach data to messages
- **Message Filtering**: Listen for specific message types

## Architecture Notes

- **Client Identity**: Each agent stores a UUID in `.needy-client-id` in their working directory
- **Server-Mediated**: All communication goes through the ndadm server (no peer-to-peer)
- **NATS Backbone**: Uses NATS for reliable, distributed messaging
- **Impersonation Prevention**: Server validates client IDs on registration
