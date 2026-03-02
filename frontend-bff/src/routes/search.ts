import { Router } from 'express';
import { getCuriosClient, getNetworkClient, call } from '../grpc/clients';
import * as grpc from '@grpc/grpc-js';

const router = Router();

// Local search — queries this node's curios-manager directly
router.get('/search/local', async (req, res) => {
  try {
    const client = getCuriosClient();
    const result = await call(client, 'ListCurios', {
      query: req.query.q ?? '',
      media_type: req.query.mediaType ?? '',
      limit: Number(req.query.limit ?? 50),
      offset: 0,
    });
    res.json(result);
  } catch (err) {
    res.status(500).json({ error: String(err) });
  }
});

// Network search — fans out to all peer nodes via network-manager (streaming)
router.get('/search/network', async (req, res) => {
  try {
    const client = getNetworkClient();
    const stream = (client as grpc.Client & {
      SearchNetwork: (req: object) => grpc.ClientReadableStream<unknown>;
    }).SearchNetwork({
      query: req.query.q ?? '',
      media_type: req.query.mediaType ?? '',
      include_unavailable: req.query.includeUnavailable === 'true',
      user_jwt: req.session.token ?? '',
    });

    const results: unknown[] = [];
    stream.on('data', (chunk: unknown) => results.push(chunk));
    stream.on('end', () => res.json({ results }));
    stream.on('error', (err: Error) => res.status(500).json({ error: err.message }));
  } catch (err) {
    res.status(500).json({ error: String(err) });
  }
});

export default router;
