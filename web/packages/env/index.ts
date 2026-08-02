import { z } from 'zod';

export const serverSchema = z.object({
  API_INTERNAL_BASE_URL: z.string().url('API_INTERNAL_BASE_URL must be a valid URL').default('http://localhost:8080'),
  NEXTAUTH_URL: z.string().url('NEXTAUTH_URL must be a valid URL').default('http://localhost:3000'),
  NEXTAUTH_SECRET: z.string().min(32, 'NEXTAUTH_SECRET must be at least 32 characters long').default('zippyra_nextauth_secret_minimum_32_chars_long'),
  ADMIN_2FA_ISSUER: z.string().default('Zippyra'),
  CLICKHOUSE_READ_HOST: z.string().default('localhost'),
  CLICKHOUSE_READ_PORT: z.union([z.string(), z.number()]).transform((val) => Number(val)).default(8123),
  CLICKHOUSE_READ_USER: z.string().default('readonly'),
  CLICKHOUSE_READ_PASSWORD: z.string().optional(),
  SENTRY_AUTH_TOKEN: z.string().optional(),
  ALLOWED_ADMIN_EMAIL_DOMAIN: z.string().default('zippyra.com'),
});

export const clientSchema = z.object({
  NEXT_PUBLIC_APP_NAME: z.enum(['retailer', 'admin', 'hq']).default('retailer'),
  NEXT_PUBLIC_ENV: z.enum(['development', 'staging', 'production']).default('development'),
  NEXT_PUBLIC_API_BASE_URL: z.string().url('NEXT_PUBLIC_API_BASE_URL must be a valid URL').default('http://localhost:8080'),
  NEXT_PUBLIC_WS_URL: z.string().default('ws://localhost:8080/ws'),
  NEXT_PUBLIC_GOOGLE_OAUTH_CLIENT_ID: z.string().optional(),
  NEXT_PUBLIC_RAZORPAY_KEY_ID: z.string().optional(),
  NEXT_PUBLIC_GOOGLE_MAPS_API_KEY: z.string().optional(),
  NEXT_PUBLIC_MAPBOX_TOKEN: z.string().optional(),
  NEXT_PUBLIC_SENTRY_DSN: z.string().optional(),
  NEXT_PUBLIC_CSP_CONNECT_SRC: z.string().default('https://api.zippyra.com'),
  NEXT_PUBLIC_ERP_INTEGRATION_DOCS_URL: z.string().url().default('https://docs.zippyra.com/erp'),
});

export type ServerEnv = z.infer<typeof serverSchema>;
export type ClientEnv = z.infer<typeof clientSchema>;

export function validateEnv(processEnv: Record<string, string | undefined> = process.env) {
  const isServer = typeof window === 'undefined';
  const environment = processEnv.NODE_ENV || processEnv.NEXT_PUBLIC_ENV || 'development';

  // Guard against secret leaks in client schema
  assertNoSecretLeaks(clientSchema);

  const serverResult = isServer ? serverSchema.safeParse(processEnv) : { success: true, data: {} as ServerEnv, error: null };
  const clientResult = clientSchema.safeParse(processEnv);

  if (!serverResult.success || !clientResult.success) {
    const serverErrors = !serverResult.success ? serverResult.error?.flatten().fieldErrors : {};
    const clientErrors = !clientResult.success ? clientResult.error?.flatten().fieldErrors : {};
    const combinedErrors = { ...serverErrors, ...clientErrors };

    const formattedMessage = Object.entries(combinedErrors)
      .map(([key, errors]) => `  - ${key}: ${errors?.join(', ')}`)
      .join('\n');

    throw new Error(
      `❌ Invalid or missing environment variables:\n${formattedMessage}\n` +
      `Please check your .env file or environment configuration.`
    );
  }

  // Production runtime checks
  if (environment === 'production' && isServer) {
    const secret = processEnv.NEXTAUTH_SECRET;
    if (!secret || secret.length < 32) {
      throw new Error('❌ NEXTAUTH_SECRET must be present and >= 32 characters in production builds.');
    }
  }

  return {
    ...(isServer ? (serverResult as any).data : {}),
    ...(clientResult as any).data,
  } as ServerEnv & ClientEnv;
}

export function assertNoSecretLeaks(clientSchemaObj: typeof clientSchema) {
  const clientKeys = Object.keys(clientSchemaObj.shape);
  const forbiddenSubstrings = ['SECRET', 'KEY_SECRET', 'PASSWORD', 'PRIVATE_KEY'];

  for (const key of clientKeys) {
    for (const sub of forbiddenSubstrings) {
      if (key.toUpperCase().includes(sub) && !key.includes('NEXT_PUBLIC_RAZORPAY_KEY_ID')) {
        throw new Error(`❌ Security Violation: Secret-sounding variable '${key}' found in client-exposed env schema!`);
      }
    }
  }
}

export const env = validateEnv();
