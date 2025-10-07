const path = require('path');

module.exports = (config, options) => {
  config.entry.background = path.join(__dirname, 'src/scripts/background.ts');
  config.entry.content = path.join(__dirname, 'src/scripts/content.ts');

  return config;
};
