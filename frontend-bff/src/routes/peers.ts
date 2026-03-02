import { Router, Request, Response, NextFunction } from 'express';
import { getNetworkClient, call } from '../grpc/clients';

const router = Router();

function requireManager(req: Request, res: Response, next: NextFunction) {
  const nodeId = process.env.NODE_ID ?? '';
  const isManager = (req.session.claims ?? []).some(
    (c) => c.node === nodeId && c.role === 'MANAGER',
  );
  if (!isManager) { res.status(403).json({ error: 'manager claim required' }); return; }
  next();
}

router.get('/peers', requireManager, async (req, res) => {
  try {
    const client = getNetworkClient();
    const result = await call(client, 'ListPeers', {});
    res.json(result);
  } catch (err) {
    res.status(500).json({ error: String(err) });
  }
});

router.post('/peers', requireManager, async (req, res) => {
  try {
    const client = getNetworkClient();
    const result = await call(client, 'RegisterPeer', {
      node_id: req.body.nodeId,
      public_key: req.body.publicKey,
      address: req.body.address,
      display_name: req.body.displayName ?? '',
    });
    res.status(201).json(result);
  } catch (err) {
    res.status(400).json({ error: String(err) });
  }
});

export default router;
