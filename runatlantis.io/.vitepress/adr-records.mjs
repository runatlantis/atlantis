import { readdirSync } from 'node:fs';

export function recordFiles(directory) {
  return readdirSync(directory).filter((file) => file === 'README.md' || /^\d{4}-.+\.md$/.test(file)).sort();
}
