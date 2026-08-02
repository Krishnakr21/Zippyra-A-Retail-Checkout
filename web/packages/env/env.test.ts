import { validateEnv, assertNoSecretLeaks, clientSchema } from './index';

describe('@zippyra/env Validation Tests', () => {
  it('validates correct environment variables without errors', () => {
    const validEnv = {
      API_INTERNAL_BASE_URL: 'http://localhost:8080',
      NEXTAUTH_URL: 'http://localhost:3000',
      NEXTAUTH_SECRET: 'zippyra_nextauth_secret_minimum_32_chars_long',
      NEXT_PUBLIC_APP_NAME: 'retailer',
      NEXT_PUBLIC_ENV: 'development',
      NEXT_PUBLIC_API_BASE_URL: 'http://localhost:8080',
      NEXT_PUBLIC_WS_URL: 'ws://localhost:8080/ws',
    };

    const parsed = validateEnv(validEnv);
    expect(parsed.NEXT_PUBLIC_API_BASE_URL).toBe('http://localhost:8080');
    expect(parsed.NEXT_PUBLIC_APP_NAME).toBe('retailer');
  });

  it('fails build if NEXT_PUBLIC_API_BASE_URL is invalid URL', () => {
    const invalidEnv = {
      NEXT_PUBLIC_API_BASE_URL: 'invalid-url',
    };

    expect(() => validateEnv(invalidEnv)).toThrow('Invalid or missing environment variables');
  });

  it('passes secret leak assertion for safe client keys', () => {
    expect(() => assertNoSecretLeaks(clientSchema)).not.toThrow();
  });
});
