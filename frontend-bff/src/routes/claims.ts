import { Router, Request, Response, NextFunction } from 'express';
import { getUsersClient, call } from '../grpc/clients';

const router = Router();

function requireManager(req: Request, res: Response, next: NextFunction) {
  const nodeId = process.env.NODE_ID ?? '';
  const isManager = (req.session.claims ?? []).some(
    (c) => c.node === nodeId && c.role === 'MANAGER',
  );
  if (!isManager) { res.status(403).json({ error: 'manager claim required' }); return; }
  next();
}

router.get('/claims', requireManager, async (req, res) => {
  try {
    const client = getUsersClient();
    const nodeId = process.env.NODE_ID ?? '';
    const result = await call(client, 'ListClaims', { node_id: nodeId });
    res.json(result);
  } catch (err) {
    res.status(500).json({ error: String(err) });
  }
});

router.post('/claims/grant', requireManager, async (req, res) => {
  try {
    const client = getUsersClient();
    await call(client, 'GrantClaim', {
      user_id: req.body.userId,
      node_id: req.body.nodeId ?? process.env.NODE_ID,
      role: req.body.role,
    });
    res.json({ ok: true });
  } catch (err) {
    res.status(400).json({ error: String(err) });
  }
});

router.delete('/claims/revoke', requireManager, async (req, res) => {
  try {
    const client = getUsersClient();
    await call(client, 'RevokeClaim', {
      user_id: req.body.userId,
      node_id: req.body.nodeId ?? process.env.NODE_ID,
    });
    res.json({ ok: true });
  } catch (err) {
    res.status(400).json({ error: String(err) });
  }
});

export default router;
