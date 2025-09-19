export function chromeIdentityLaunchFlow(url: string): Promise<string> {
  return new Promise((resolve, reject) => {
    chrome.identity.launchWebAuthFlow({ url, interactive: true }, redirectUrl => {
      if (chrome.runtime.lastError) {
        return reject(chrome.runtime.lastError);
      }
      resolve(redirectUrl!);
    });
  });
}

export function chromeStorageGet<T>(key: string): Promise<T | null> {
  return new Promise<T | null>(resolve => {
    chrome.storage.local.get([key], (result: Record<string, unknown>) => {
      const val = result[key];
      resolve((val !== undefined ? (val as T) : null));
    });
  });
}

export function chromeStorageSet(key: string, value: unknown): Promise<void> {
  return new Promise(resolve => {
    chrome.storage.local.set({ [key]: value }, () => resolve());
  });
}
