import { fileURLToPath } from 'node:url';
import { fetchDocumentationHistory } from './build-utils.mjs';

fetchDocumentationHistory(fileURLToPath(new URL('../../', import.meta.url)));
