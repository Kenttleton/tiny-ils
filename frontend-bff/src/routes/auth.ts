import { Router, Request, Response } from 'express';
import { generators } from 'openid-client';
import { getUsersClient, call } from '../grpc/clients';
import { getGoogleClient } from '../auth/oidc';
import { decodeJWTPayload } from '../auth/session';

const router = Router();

// ─── Password auth ────────────────────────────────────────────────────────────

router.post('/auth/register', async (req, res) => {
  try {
    const client = getUsersClient();
    const result = await call<{ token: string; user: object }>(client, 'Register', {
      email: req.body.email,
      password: req.body.password,
      display_name: req.body.displayName ?? '',
    });
    storeSession(req, result.token);
    res.json({ user: result.user });
  } catch (err: unknown) {
    res.status(400).json({ error: grpcMessage(err) });
  }
});

router.post('/auth/login', async (req, res) => {
  try {
    const client = getUsersClient();
    const result = await call<{ token: string; user: object }>(client, 'Login', {
      email: req.body.email,
      password: req.body.password,
    });
    storeSession(req, result.token);
    res.json({ user: result.user });
  } catch (err: unknown) {
    res.status(401).json({ error: grpcMessage(err) });
  }
});

router.post('/auth/logout', (req, res) => {
  req.session.destroy(() => res.json({ ok: true }));
});

router.get('/auth/me', (req, res) => {
  if (!req.session.userId) return res.status(401).json({ error: 'not authenticated' });
  res.json({ userId: req.session.userId, claims: req.session.claims ?? [] });
});

// ─── Google SSO ───────────────────────────────────────────────────────────────

router.get('/auth/google', async (req, res) => {
  const client = await getGoogleClient();
  if (!client) return res.status(501).json({ error: 'Google SSO not configured' });

  const state = generators.state();
  const nonce = generators.nonce();
  req.session.oauthState = state;
  req.session.oauthNonce = nonce;

  const url = client.authorizationUrl({
    scope: 'openid email profile',
    state,
    nonce,
  });
  res.redirect(url);
});

router.get('/auth/callback/google', async (req, res) => {
  const client = await getGoogleClient();
  if (!client) return res.status(501).json({ error: 'Google SSO not configured' });

  try {
    const params = client.callbackParams(req);
    const tokenSet = await client.callback(
      process.env.GOOGLE_REDIRECT_URI ?? 'http://localhost:3001/auth/callback/google',
      params,
      { state: req.session.oauthState, nonce: req.session.oauthNonce },
    );
    const userinfo = await client.userinfo(tokenSet);

    const usersClient = getUsersClient();
    const result = await call<{ token: string; user: object }>(usersClient, 'UpsertSSOUser', {
      provider: 'google',
      subject: userinfo.sub,
      email: userinfo.email,
      display_name: userinfo.name ?? userinfo.email,
    });
    storeSession(req, result.token);
    res.redirect(process.env.FRONTEND_URL ?? 'http://localhost:3000');
  } catch (err: unknown) {
    res.status(400).json({ error: grpcMessage(err) });
  }
});

// ─── Helpers ─────────────────────────────────────────────────────────────────

function storeSession(req: Request, token: string) {
  const payload = decodeJWTPayload(token);
  req.session.userId = payload['uid'] as string;
  req.session.token = token;
  req.session.claims = payload['claims'] as Array<{ node: string; role: string }>;
}

function grpcMessage(err: unknown): string {
  if (err && typeof err === 'object' && 'details' in err) return (err as { details: string }).details;
  return String(err);
}

// Extend session type with OAuth state fields
declare module 'express-session' {
  interface SessionData {
    oauthState?: string;
    oauthNonce?: string;
  }
}

export default router;
