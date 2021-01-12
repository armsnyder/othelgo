import { Readable, writable } from "svelte/store";
import type { InboundMessage, OutboundMessage } from "./messageTypes";

// Initialize a Svelte store. Writing to the store will notify all subscribers.
const messageStore = writable<InboundMessage | null>(null);

// Create a websocket connection.
const socket = new WebSocket(
  "wss://1y9vcb5geb.execute-api.us-west-2.amazonaws.com/development"
);

// Buffer outbound message while websocket is connecting.
const outboundQueue: OutboundMessage[] = [];

// Send buffered messages when the websocket opens.
socket.addEventListener("open", () => {
  outboundQueue.forEach((message) => {
    socket.send(JSON.stringify(message));
  });
  outboundQueue.length = 0;
});

// Receive messages from the websocket connection.
socket.addEventListener("message", ({ data }) => {
  messageStore.set(JSON.parse(data));
});

// Function for sending messages over the websocket connection.
export const sendMessage = (message: OutboundMessage) => {
  if (socket.readyState == 0) {
    outboundQueue.push(message);
  } else if (socket.readyState == 1) {
    socket.send(JSON.stringify(message));
  } else {
    throw new Error(`Websocket readystate is ${socket.readyState}`);
  }
};

// Create a readable store that receives a specific message type.
// Svelte components can use the $ shorthand to auto-subscribe to the latest value.
export const createMessageReceiver = <T extends InboundMessage>(
  action: string
): Readable<T | null> => ({
  subscribe: (run, invalidate) =>
    messageStore.subscribe((value) => {
      if (!value || value.action === action) {
        run(value as T | null);
      }
    }, invalidate),
});
