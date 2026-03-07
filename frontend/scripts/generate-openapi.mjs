import { writeFileSync } from 'fs';
import converter from 'swagger2openapi';

const SWAGGER_URL = 'https://api.piratereads.com/swagger/doc.json';

const response = await fetch(SWAGGER_URL);
if (!response.ok) {
  throw new Error(`Failed to fetch swagger spec: ${response.status}`);
}

const swaggerSpec = await response.json();
const { openapi } = await converter.convertObj(swaggerSpec, { patch: true, warnOnly: true });

writeFileSync(
  new URL('../openapi.json', import.meta.url),
  JSON.stringify(openapi, null, 2),
);

console.log('Generated openapi.json');
