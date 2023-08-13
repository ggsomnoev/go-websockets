# WebSocket Implementation with Golang Backend and Plain HTML/CSS/JavaScript Frontend

This repository demonstrates a WebSocket communication implementation using Golang as the backend and a simple HTML/CSS/JavaScript frontend.

## Frontend

### Login Form
- The login form sends user form data to the backend for validation.
- If authorized, an OTP (One-Time Password) is sent back.

### SendMessage Form
- Upon submission, a `sendmessage` event is sent to the backend.
- Assumes a valid and connected user.
- Includes message content and username.

### ChangeChatRoom Form
- Users can input a chatroom name; submission associates the chatroom with the user.

An event router validates message types and appends them to the chat message box.

## Backend

HTTPS server with three endpoints:

1. **"/"**: Renders the frontend.
2. **"/login"**: Validates users and generates OTP.
3. **"/ws"**: Creates a WebSocket connection.

Backend components:

- **WebSocket Manager**:
  - Manages WebSocket functionality.
  - Validates request origin and provided OTP.
  - Maintains a list of clients.
  - Clients send/receive messages via an egress channel.
  - Each client is associated with a chatroom and manager.
  - Limited message size to prevent large frames.
  - Ping functionality checks client connection.
  - Routes messages to appropriate handlers based on type.

Supported Event Message Types:

- **new_message (Received)**: Handles received messages.
- **send_message (Send)**: Manages outgoing messages.
- **change_room**: Deals with changing chat rooms.

## Getting Started

To run the project:

1. Run gencert.bash.
2. Run the project :)
