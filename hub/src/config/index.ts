import fs from 'fs';
import yaml from 'yaml';
import path from 'path';

const configPath = path.join(process.cwd(), 'config.yaml');
const configFile = fs.readFileSync(configPath, 'utf8');
const yamlConfig = yaml.parse(configFile);

export const config = {
  hub: {
    address: yamlConfig.hub.address || '0.0.0.0',
    address6: yamlConfig.hub.address6 || '::',
    port: yamlConfig.hub.port || 3000,
    tcpPort: yamlConfig.hub.tcpPort || 3001,
    enableIPv6: yamlConfig.hub.enableIPv6 || false,
  },
  db: {
    path: yamlConfig.db.path || path.join(process.cwd(), 'data/hub.db'),
  },
  log: {
    level: yamlConfig.log.level || 'info',
    path: yamlConfig.log.path || 'logs/hub.log',
    timeFormat: yamlConfig.log.timeFormat || 'YYYY-MM-DD HH:mm:ss.SSS',
  },
  auth: {
    key: yamlConfig.auth.key || 'default-key-not-secure',
  },
};

export { config }