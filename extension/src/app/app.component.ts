import { Component, OnInit } from '@angular/core';
import { RouterOutlet } from '@angular/router';
import { HeaderComponent } from './theme/components/header/header.component';
import { FooterComponent } from './theme/components/footer/footer.component';
import { InvitePopupComponent } from "theme/components/popups/invite/invite-popup.component";

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [RouterOutlet, HeaderComponent, FooterComponent, InvitePopupComponent],
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.scss'],
})
export class AppComponent implements OnInit {
  ngOnInit() {
    this.checkCurrentTab();
  }

  private async checkCurrentTab() {
    if (typeof chrome === "undefined" || !chrome.tabs) {
      return; // We are not in a popup, probably in the iframe or new tab
    }
    
    const [tab] = await chrome.tabs.query({
      active: true,
      currentWindow: true
    });

    if (tab && tab.id && tab.url && tab.url.includes("https://leetcode.com/problems/")) {
      chrome.tabs.sendMessage(tab.id, { action: "toggle_ui" }, () => {
        if (chrome.runtime.lastError) {
          console.error(chrome.runtime.lastError.message);
        }
        
        window.close();
      });
    }
  }
}