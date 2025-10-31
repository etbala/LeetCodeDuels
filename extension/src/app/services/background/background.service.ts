import { Injectable } from '@angular/core';
import { 
  AcceptInvitationPayload,
  BackgroundAction,
  BackgroundActionType,
  DeclineInvitationPayload,
  SendInvitationPayload
} from '../../models/background-actions';

@Injectable({
  providedIn: 'root',
})
export class BackgroundService {

  private sendMessage<T, P>(action: BackgroundActionType, payload?: P): Promise<T> {
    return new Promise((resolve, reject) => {
      const message: BackgroundAction<P> = { action, payload };

      chrome.runtime.sendMessage(message, (response) => {
        if (chrome.runtime.lastError) {
          return reject(chrome.runtime.lastError.message);
        }
        if (response?.status === 'error') {
          return reject(response.error);
        }
        resolve(response as T);
      });
    });
  }

  sendInvitation(payload: SendInvitationPayload) {
    return this.sendMessage(BackgroundActionType.DuelSendInvitation, payload);
  }

  acceptInvitation(payload: AcceptInvitationPayload) {
    return this.sendMessage(BackgroundActionType.DuelAcceptInvitation, payload);
  }

  declineInvitation(payload: DeclineInvitationPayload) {
    return this.sendMessage(BackgroundActionType.DuelDeclineInvitation, payload);
  }

  cancelInvitation() {
    return this.sendMessage(BackgroundActionType.DuelCancelInvitation);
  }

  forfeitDuel() {
    return this.sendMessage(BackgroundActionType.DuelForfeit);
  }
}