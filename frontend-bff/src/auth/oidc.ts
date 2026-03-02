import { Issuer, generators } from 'openid-client';
import type { Client } from 'openid-client';

let googleClient: Client | null = null;

export async function getGoogleClient(): Promise<Client | null> {
  const clientId = process.env.GOOGLE_CLIENT_ID;
  const clientSecret = process.env.GOOGLE_CLIENT_SECRET;
  if (!clientId || !clientSecret) return null;

  if (!googleClient) {
    const issuer = await Issuer.discover('https://accounts.google.com');
    googleClient = new issuer.Client({
      client_id: clientId,
      client_secret: clientSecret,
      redirect_uris: [process.env.GOOGLE_REDIRECT_URI ?? 'http://localhost:3001/auth/callback/google'],
      response_types: ['code'],
    });
  }
  return googleClient;
}

export { generators };
