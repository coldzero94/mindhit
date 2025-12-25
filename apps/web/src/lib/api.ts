import { createClient } from '../api/generated';

export const apiClient = createClient({
  baseUrl: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080',
});

// Re-export SDK functions for convenience
export * from '../api/generated/sdk.gen';

// Re-export Zod schemas for validation
export * from '../api/generated/zod.gen';

// Re-export types
export type * from '../api/generated/types.gen';
