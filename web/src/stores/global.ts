// Globally-shared variables.

import { writable } from "svelte/store";
import { createPersistentStore } from "./localStorage";

// host is the host player's nickname.
// If the client user is hosting the game, it will be equal to their nickname.
// If the client user is not currently in-game, it will be empty.
export const host = writable<string>("");

type SceneName = "join" | "nickname" | "game";

// currentScene can be written to in order to trigger a scene change.
export const currentScene = writable<SceneName>("join");

// nickname is the client user's nickname.
export const nickname = createPersistentStore("nickname");
