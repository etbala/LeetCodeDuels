export interface BackgroundScriptMessage {
  action: BackgroundAction;
  payload?: any;
}

export type BackgroundAction =
  | 'ws:connect'
  | 'ws:disconnect'
  | 'duel:send-invitation'
  | 'duel:accept-invitation'
  | 'duel:decline-invitation'
  | 'duel:cancel-invitation'
//   | 'queue:enter'
//   | 'queue:leave'
  | 'duel:submission';

export interface ExtensionEvent {
  event: string;
  data: any;
}