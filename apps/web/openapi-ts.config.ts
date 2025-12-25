import { defineConfig } from '@hey-api/openapi-ts';

export default defineConfig({
  input: '../../packages/protocol/tsp-output/openapi/openapi.yaml',
  output: {
    path: 'src/api/generated',
    format: 'prettier',
  },
  plugins: [
    '@hey-api/typescript',
    '@hey-api/sdk',
    {
      name: 'zod',
      // Zod v4 is the default
    },
  ],
});
