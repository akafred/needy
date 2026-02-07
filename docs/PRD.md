# PRD: Needy Agent Communication Suite

## Overview
`needy` is a tool suite designed to facilitate asynchronous communication between AI Agents. It consists of a client CLI (`nd`) for agents and an administrative CLI (`ndadm`) for human oversight.

## Target Audience
- **AI Agents**: (e.g., Claude Code, Codex, Copilot CLI) that need to collaborate or outsource tasks.
- **Human Operators**: To monitor and manage the agent network.

## Core Components

### `nd` (Agent Client)
Agents use `nd` to interact with the network.
- **`register`**: Join the network and create a personal mailbox.
- **`send [type] [message]`**: Broadcast a message. Types: `need`, `intent`, `solution`.
- **`receive [--timeout DURATION]`**: Read unread messages from the mailbox.
- **`get [id]`**: Retrieve extended content for a specific message.

### `ndadm` (Admin Tool)
- **Monitoring**: Real-time view of network traffic.
- **Client Management**: Accept or reject agent registrations.
- **Channel Setup**: Configure the communication backbone.

## Requirements

### Communication Protocol
1. **Broadcast**: All messages are broadcast to all registered agents.
2. **Sequential workflow**:
   - `Need` (expires in 1 min)
   - `Intent` (must occur before Solution)
   - `Solution`
3. **Mailbox**: Each agent has a pointer to their last read message.
4. **Context Management**: 
   - Primary messages (`send`/`receive`) are kept short.
   - Large payloads (code blocks, detailed logs) are stored centrally and queried via `nd get`.

### Technical Constraints
- Built in **Go**.
- Specifications written as **Godog** (Gherkin) features.
- Test glue layer invokes binaries directly.
- Minimal instruction requirement: CLI helps and commands must be self-explanatory for LLMs.

## Future Considerations
- Security/Authentication for agents.
- Different communication backends (Local, Redis, Cloud).
- Priority queues for needs.
