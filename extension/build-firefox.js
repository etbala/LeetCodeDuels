const fs = require('fs');
const { execSync } = require('child_process');

let config = 'dev';

if (process.env.npm_config_config) {
  // This is populated by running `npm run ... --config=local`
  config = process.env.npm_config_config;
  console.log(`Found config from --config argument: ${config}`);
} else if (process.env.npm_config_configuration) {
  // This is populated by running `npm run ... --configuration=local`
  config = process.env.npm_config_configuration;
  console.log(`Found config from --configuration argument: ${config}`);
} else {
  console.log(`No --config or --configuration argument found in env. Defaulting to "${config}".`);
}

fs.copyFileSync('src/manifest.json', 'src/manifest.json.backup');
fs.copyFileSync('src/manifest.firefox.json', 'src/manifest.json');

try {
  execSync(`ng build --configuration=${config}`, { stdio: 'inherit' });
} finally {
  fs.copyFileSync('src/manifest.json.backup', 'src/manifest.json');
  fs.unlinkSync('src/manifest.json.backup');
}