import session from 'express-session';

export interface SessionData {
  userId?: string;
  token?: string;  // node-signed JWT (stored server-side only)
  claims?: Array<{ node: string; role: string }>;
}

declare module 'express-session' {
  interface SessionData {
    userId?: string;
    token?: string;
    claims?: Array<{ node: string; role: string }>;
  }
}

export function sessionMiddleware() {
  const secret = process.env.SESSION_SECRET ?? 'changeme-set-SESSION_SECRET-in-env';
  return session({
    secret,
    resave: false,
    saveUninitialized: false,
    cookie: {
      httpOnly: true,
      sameSite: 'lax',
      secure: process.env.NODE_ENV === 'production',
      maxAge: 24 * 60 * 60 * 1000, // 24 hours
    },
  });
}

/** Decode a JWT payload without verifying the signature. Used only to read claims. */
export function decodeJWTPayload(token: string): Record<string, unknown> {
  const parts = token.split('.');
  if (parts.length !== 3) throw new Error('Invalid JWT');
  const payload = Buffer.from(parts[1], 'base64url').toString('utf8');
  return JSON.parse(payload);
}
