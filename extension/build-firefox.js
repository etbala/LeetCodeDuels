const fs = require('fs');
const { execSync } = require('child_process');

fs.copyFileSync('src/manifest.json', 'src/manifest.json.backup');
fs.copyFileSync('src/manifest.firefox.json', 'src/manifest.json');

try {
  execSync('ng build --configuration=firefox', { stdio: 'inherit' });
} finally {
  fs.copyFileSync('src/manifest.json.backup', 'src/manifest.json');
  fs.unlinkSync('src/manifest.json.backup');
}