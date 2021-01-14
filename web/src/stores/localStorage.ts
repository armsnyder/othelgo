import { writable, Writable } from "svelte/store";

// createStore creates a new Svelte store that reads and saves values
// to to the client browser's localStorage.
export const createStore = (key: string): Writable<string> => {
  const store = writable(localStorage.getItem(key) ?? "");
  store.subscribe((value) => localStorage.setItem(key, value));
  return store;
};
