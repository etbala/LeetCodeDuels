interface Window {
  __CODE_DUELS_EXTENSION_INJECTED__?: boolean;
}

function main() {
  console.log("Content script injected.");
}

if (window.__CODE_DUELS_EXTENSION_INJECTED__) {
  console.log("Content script already injected, stopping.");
} else {
  window.__CODE_DUELS_EXTENSION_INJECTED__ = true;
  main();
}