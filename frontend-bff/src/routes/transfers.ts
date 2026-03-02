import { Router, Request, Response, NextFunction } from 'express';
import { getCuriosClient, getNetworkClient, call } from '../grpc/clients';

const router = Router();

function requireAuth(req: Request, res: Response, next: NextFunction) {
  if (!req.session.userId) { res.status(401).json({ error: 'not authenticated' }); return; }
  next();
}

// ─── List transfers ────────────────────────────────────────────────────────────

router.get('/transfers', requireAuth, async (req, res) => {
  try {
    const client = getCuriosClient();
    const result = await call(client, 'ListTransfers', {
      status: req.query.status ?? '',
      node_id: req.query.nodeId ?? '',
      transfer_type: req.query.transferType ?? '',
    });
    res.json(result);
  } catch (err) {
    res.status(500).json({ error: String(err) });
  }
});

// ─── Request a transfer ────────────────────────────────────────────────────────
// If source_node matches the local node we go directly to curios-manager.
// If source_node is a different node we forward via network-manager.

router.post('/transfers', requireAuth, async (req, res) => {
  try {
    const nodeId = process.env.NODE_ID ?? '';
    const { copyId, transferType, sourceNode, destNode, notes } = req.body as {
      copyId: string;
      transferType: string;
      sourceNode: string;
      destNode: string;
      notes?: string;
    };

    if (sourceNode === nodeId || !sourceNode || sourceNode === '') {
      // Local transfer — the copy lives here
      const client = getCuriosClient();
      const result = await call(client, 'RequestTransfer', {
        copy_id: copyId,
        transfer_type: transferType,
        source_node: sourceNode || nodeId,
        dest_node: destNode,
        initiated_by: req.session.userId,
        notes: notes ?? '',
      });
      res.status(201).json(result);
    } else {
      // Remote transfer — forward to the source node via network-manager
      const netClient = getNetworkClient();
      const ack = await call(netClient, 'InitiateRemoteTransfer', {
        transfer_id: '',
        copy_id: copyId,
        transfer_type: transferType,
        source_node: sourceNode,
        dest_node: destNode || nodeId,
        initiated_by: req.session.userId,
        user_jwt: req.session.token ?? '',
        notes: notes ?? '',
      });
      res.status(201).json(ack);
    }
  } catch (err) {
    res.status(400).json({ error: String(err) });
  }
});

// ─── Get a transfer ────────────────────────────────────────────────────────────

router.get('/transfers/:id', requireAuth, async (req, res) => {
  try {
    const client = getCuriosClient();
    const result = await call(client, 'GetTransfer', { id: req.params.id });
    res.json(result);
  } catch (err) {
    res.status(404).json({ error: String(err) });
  }
});

// ─── Transfer actions ──────────────────────────────────────────────────────────

router.post('/transfers/:id/approve', requireAuth, async (req, res) => {
  try {
    const client = getCuriosClient();
    const result = await call(client, 'ApproveTransfer', {
      transfer_id: req.params.id,
      actor_id: req.session.userId,
      notes: req.body.notes ?? '',
    });
    res.json(result);
  } catch (err) {
    res.status(400).json({ error: String(err) });
  }
});

router.post('/transfers/:id/reject', requireAuth, async (req, res) => {
  try {
    const client = getCuriosClient();
    const result = await call(client, 'RejectTransfer', {
      transfer_id: req.params.id,
      actor_id: req.session.userId,
      notes: req.body.notes ?? '',
    });
    res.json(result);
  } catch (err) {
    res.status(400).json({ error: String(err) });
  }
});

router.post('/transfers/:id/ship', requireAuth, async (req, res) => {
  try {
    const client = getCuriosClient();
    const result = await call(client, 'MarkShipped', {
      transfer_id: req.params.id,
      actor_id: req.session.userId,
      notes: req.body.notes ?? '',
    });
    res.json(result);
  } catch (err) {
    res.status(400).json({ error: String(err) });
  }
});

router.post('/transfers/:id/receive', requireAuth, async (req, res) => {
  try {
    const client = getCuriosClient();
    const result = await call(client, 'ConfirmReceived', {
      transfer_id: req.params.id,
      actor_id: req.session.userId,
      notes: req.body.notes ?? '',
    });
    res.json(result);
  } catch (err) {
    res.status(400).json({ error: String(err) });
  }
});

router.post('/transfers/:id/cancel', requireAuth, async (req, res) => {
  try {
    const client = getCuriosClient();
    const result = await call(client, 'CancelTransfer', {
      transfer_id: req.params.id,
      actor_id: req.session.userId,
      notes: req.body.notes ?? '',
    });
    res.json(result);
  } catch (err) {
    res.status(400).json({ error: String(err) });
  }
});

export default router;
