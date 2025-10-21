import { Injectable, NgZone } from '@angular/core';
import { Observable, Subject } from 'rxjs';
import { filter, map } from 'rxjs/operators';
import { ExtensionEventType, UIMessage } from 'models/extension-events';

@Injectable({
  providedIn: 'root',
})
export class ExtensionEventsService {
  private messageSubject = new Subject<UIMessage<unknown>>();

  constructor(private ngZone: NgZone) {
    this.listenForMessages();
  }

  /**
   * Sets up the main listener for all messages from the service worker.
   */
  private listenForMessages(): void {
    chrome.runtime.onMessage.addListener((message: UIMessage<unknown>, sender, sendResponse) => {
      // Messages from the service worker won't have a tab sender.
      if (sender.tab) {
        return;
      }
      
      // Run inside NgZone to ensure change detection is triggered
      // in components that subscribe to these events.
      this.ngZone.run(() => {
        this.messageSubject.next(message);
      });
    });
  }

  /**
   * A generic method for components to subscribe to specific event types.
   * @param type The ExtensionEventType to listen for.
   * @returns An Observable that emits the payload of messages with the matching type.
   */
  public listenFor<T>(type: ExtensionEventType): Observable<T> {
    return this.messageSubject.asObservable().pipe(
      filter((message) => message.type === type),
      map((message) => message.payload as T)
    );
  }
}