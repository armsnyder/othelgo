import { writable } from "svelte/store";

// Initialize a Svelte store. Writing to the store will notify all subscribers.
const messageStore = writable(null);

// Create a websocket connection.
const socket = new WebSocket(
  "wss://1y9vcb5geb.execute-api.us-west-2.amazonaws.com/development"
);

// Receive messages from the websocket connection.
socket.addEventListener("message", ({ data }) => {
  messageStore.set(JSON.parse(data));
});

// Function for sending messages over the websocket connection.
export const sendMessage = (message) => {
  socket.send(JSON.stringify(message));
};

// This message object conforms to the Svelte Store interface, meaning that Svelte components can
// use the $ shorthand to auto-subscribe to the latest value. The value will be the most recently
// received message from the websocket connection.
export const message = {
  subscribe: messageStore.subscribe,
};
