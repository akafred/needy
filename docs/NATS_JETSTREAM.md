# NATS JetStream Integration in Needy

This document explains how `nd` and `ndadm` use NATS JetStream for reliable messaging.

## High-Level Concept

Think of **JetStream** as a persistent log of messages (like a Kafka topic or a persistent mailbox) that lives inside the `ndadm` server.

Without JetStream (standard NATS), if an agent is offline when a message is sent, that message is gone forever (fire-and-forget).
**With JetStream**, messages are stored on disk, and agents can "catch up" on what they missed.

## Architecture

### 1. The Stream (`MESSAGES`)
On startup, `ndadm` creates a **Stream** called `MESSAGES`.
- **Subject**: `needy.messages`
- **Storage**: File-based (saved to `.nats-data/` directory)
- **Retention**: Currently configured to keep messages forever (default).

Every time an agent sends a Need, Intent, or Solution, `ndadm` publishes it to this stream.

### 2. Durable Consumers (Mailboxes)
When an agent runs `nd receive`, `ndadm` creates (or reuses) a **Durable Consumer** for that agent.
- **Consumer Name**: `AGENT_<AgentName>` (e.g., `AGENT_AgentAlice`)
- **Function**: Keeps a "bookmark" of the last message THIS agent successfully processed.

This allows independent reading:
- AgentAlice might be on Message #5.
- AgentBob might be on Message #20.
- If AgentAlice crashes and restarts, the Consumer remembers she is at #5, so she gets #6 next.

### 3. The Flow

#### Sending (`nd send`)
1. Client sends JSON payload to `needy.send` (a request/reply subject).
2. `ndadm` validates the request (checks client ID, intent rules).
3. `ndadm` **publishes** the valid message to the JetStream subject `needy.messages`.
4. JetStream writes it to disk and assigns it a sequence number (ID).
5. `ndadm` replies "Success" to the client.

#### Receiving (`nd receive`)
1. Client sends a request to `needy.read`.
2. `ndadm` looks up the Durable Consumer for that agent.
3. `ndadm` asks JetStream: "Give me the next 10 messages for `AGENT_AgentAlice`".
4. JetStream returns messages starting from the agent's bookmark.
5. `ndadm` sends them to the client.
6. The exact message ID (Sequence Number) is used to track progress.

## Why this matters?
- **Persistence**: You can kill `ndadm`, delete the binary, rebuild it, and if `.nats-data` is preserved, all message history is safe.
- **Reliability**: Agents don't need to be online simultaneously to communicate.
- **Decoupling**: Senders don't need to know who is listening.
