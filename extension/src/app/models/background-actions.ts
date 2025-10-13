import { MatchDetails } from "./match";

// All possible event types sent from angular UI to the background script.
export enum BackgroundActionType {
  WSConnect = 'ws:connect',
  WSDisconnect = 'ws:disconnect',
  DuelSendInvitation = 'duel:send-invitation',
  DuelAcceptInvitation = 'duel:accept-invitation',
  DuelDeclineInvitation = 'duel:decline-invitation',
  DuelCancelInvitation = 'duel:cancel-invitation',
  DuelSubmission = 'duel:submission',
  // QueueEnter = 'queue:enter',
  // QueueLeave = 'queue:leave',
}

export interface SendInvitationPayload {
  inviteeID: number;
  matchDetails: MatchDetails;
}

export interface AcceptInvitationPayload {
  inviterID: number;
}

export interface DeclineInvitationPayload {
  inviterID: number;
}

export interface BackgroundAction<T = string> {
  action: BackgroundActionType;
  payload?: T;
}
