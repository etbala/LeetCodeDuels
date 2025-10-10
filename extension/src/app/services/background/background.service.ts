import { Injectable } from '@angular/core';
import { BackgroundAction, BackgroundScriptMessage } from '../../models/background-actions';

@Injectable({
  providedIn: 'root',
})
export class BackgroundService {

  private sendMessage<T>(action: BackgroundAction, payload?: any): Promise<T> {
    return new Promise((resolve, reject) => {
      const message: BackgroundScriptMessage = { action, payload };
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

  sendInvitation(payload: { inviteeId: string; problemId: string }) {
    return this.sendMessage('duel:send-invitation', payload);
  }

  acceptInvitation(payload: { invitationId: string }) {
    return this.sendMessage('duel:accept-invitation', payload);
  }

  declineInvitation(payload: { invitationId: string }) {
    return this.sendMessage('duel:decline-invitation', payload);
  }

  cancelInvitation(payload: { invitationId: string }) {   
    return this.sendMessage('duel:cancel-invitation', payload);
  }

}