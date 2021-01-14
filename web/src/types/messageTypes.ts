import type { Board, Player } from "./boardTypes";

export type OutboundMessage =
  | Hello
  | HostGame
  | StartSoloGame
  | JoinGame
  | LeaveGame
  | ListOpenGames
  | PlaceDisk;

export type InboundMessage =
  | Joined
  | GameOver
  | OpenGames
  | UpdateBoard
  | Error
  | Decorate;

export interface Hello {
  action: "hello";
  version: string;
}

export interface HostGame {
  action: "hostGame";
  nickname: string;
}

export interface StartSoloGame {
  action: "startSoloGame";
  nickname: string;
  difficulty: number;
}

export interface JoinGame {
  action: "joinGame";
  nickname: string;
  host: string;
}

export interface Joined {
  action: "joined";
  nickname: string;
}

export interface LeaveGame {
  action: "leaveGame";
  nickname: string;
  host: string;
}

export interface GameOver {
  action: "gameOver";
  message: string;
}

export interface ListOpenGames {
  action: "listOpenGames";
}

export interface OpenGames {
  action: "openGames";
  hosts: string[];
}

export interface PlaceDisk {
  action: "placeDisk";
  nickname: string;
  host: string;
  x: number;
  y: number;
}

export interface UpdateBoard {
  action: "updateBoard";
  board: Board;
  player: Player;
  x: number;
  y: number;
  p1score: number;
  p2score: number;
}

export interface Error {
  action: "error";
  error: string;
}

export interface Decorate {
  action: "decorate";
  decoration: string;
}
