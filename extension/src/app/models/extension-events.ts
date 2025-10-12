import { MatchDetails, PlayerSubmission } from "./match";

// All possible event types sent from the background script to the UI.
export enum ExtensionEventType {
  InvitationRequest = "invitation_request",
  StartGame = "start_game",
  GameOver = "game_over",
  OpponentSubmission = "opponent_submission",
  InvitationCanceled = "invitation_canceled",
  InvitationDeclined = "invitation_declined",
  UserOffline = "user_offline",
  InvitationNonexistant = "invitation_nonexistant",
  OtherLogon = "other_logon",
  ServerError = "server_error",
}

export interface UIMessage<T> {
  type: ExtensionEventType;
  payload: T;
}

// --- Payload Type Definitions ---

export interface StartGamePayload {
  sessionID: string;
  problemURL: string;
  opponentID: number;
}

export interface InvitationRequestPayload {
  inviterID: number;
  matchDetails: MatchDetails;
}

export interface GameOverPayload {
  winnerID: number;
  sessionID: string;
  duration: number;
}

export type OpponentSubmissionPayload = PlayerSubmission;

// Payloads for events that just carry an ID
export interface UserIdentifiable {
  inviterID?: number;
  inviteeID?: number;
}