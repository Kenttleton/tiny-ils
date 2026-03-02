import { Router, Request, Response, NextFunction } from 'express';
import multer from 'multer';
import { createHash } from 'crypto';
import { getCuriosClient, call } from '../grpc/clients';

const router = Router();
const upload = multer({ storage: multer.memoryStorage(), limits: { fileSize: 500 * 1024 * 1024 } });

function requireAuth(req: Request, res: Response, next: NextFunction) {
  if (!req.session.userId) { res.status(401).json({ error: 'not authenticated' }); return; }
  next();
}

// ─── Catalog ──────────────────────────────────────────────────────────────────

router.get('/curios', async (req, res) => {
  try {
    const client = getCuriosClient();
    const result = await call(client, 'ListCurios', {
      query: req.query.q ?? '',
      media_type: req.query.mediaType ?? '',
      format_type: req.query.formatType ?? '',
      tags: req.query.tags ? String(req.query.tags).split(',') : [],
      limit: Number(req.query.limit ?? 50),
      offset: Number(req.query.offset ?? 0),
    });
    res.json(result);
  } catch (err) {
    res.status(500).json({ error: String(err) });
  }
});

router.get('/curios/:id', async (req, res) => {
  try {
    const client = getCuriosClient();
    const result = await call(client, 'GetCurio', { id: req.params.id });
    res.json(result);
  } catch (err) {
    res.status(404).json({ error: String(err) });
  }
});

router.post('/curios', requireAuth, async (req, res) => {
  try {
    const client = getCuriosClient();
    const result = await call(client, 'CreateCurio', req.body);
    res.status(201).json(result);
  } catch (err) {
    res.status(400).json({ error: String(err) });
  }
});

router.put('/curios/:id', requireAuth, async (req, res) => {
  try {
    const client = getCuriosClient();
    const result = await call(client, 'UpdateCurio', { id: req.params.id, ...req.body });
    res.json(result);
  } catch (err) {
    res.status(400).json({ error: String(err) });
  }
});

router.delete('/curios/:id', requireAuth, async (req, res) => {
  try {
    const client = getCuriosClient();
    await call(client, 'DeleteCurio', { id: req.params.id });
    res.status(204).send();
  } catch (err) {
    res.status(400).json({ error: String(err) });
  }
});

// ─── Metadata enrichment ──────────────────────────────────────────────────────

router.post('/curios/enrich', requireAuth, async (req, res) => {
  try {
    const client = getCuriosClient();
    const result = await call(client, 'EnrichMetadata', {
      media_type: req.body.mediaType,
      identifier: req.body.identifier,
    });
    res.json(result);
  } catch (err) {
    res.status(400).json({ error: String(err) });
  }
});

// ─── Physical copies and loans ────────────────────────────────────────────────

router.get('/curios/:id/copies', async (req, res) => {
  try {
    const client = getCuriosClient();
    const result = await call(client, 'ListCopies', { id: req.params.id });
    res.json(result);
  } catch (err) {
    res.status(500).json({ error: String(err) });
  }
});

router.post('/copies/:id/checkout', requireAuth, async (req, res) => {
  try {
    const client = getCuriosClient();
    const result = await call(client, 'CheckoutCopy', {
      copy_id: req.params.id,
      user_id: req.session.userId,
      user_node_id: req.body.userNodeId ?? '',
      due_date: req.body.dueDate ?? 0,
    });
    res.json(result);
  } catch (err) {
    res.status(400).json({ error: String(err) });
  }
});

router.post('/copies/:id/return', requireAuth, async (req, res) => {
  try {
    const client = getCuriosClient();
    const result = await call(client, 'ReturnCopy', { copy_id: req.params.id });
    res.json(result);
  } catch (err) {
    res.status(400).json({ error: String(err) });
  }
});

router.post('/curios/:id/hold', requireAuth, async (req, res) => {
  try {
    const client = getCuriosClient();
    const result = await call(client, 'PlaceHold', {
      curio_id: req.params.id,
      user_id: req.session.userId,
      user_node_id: req.body.userNodeId ?? '',
    });
    res.json(result);
  } catch (err) {
    res.status(400).json({ error: String(err) });
  }
});

router.delete('/holds/:id', requireAuth, async (req, res) => {
  try {
    const client = getCuriosClient();
    await call(client, 'CancelHold', { id: req.params.id });
    res.status(204).send();
  } catch (err) {
    res.status(400).json({ error: String(err) });
  }
});

// ─── Digital assets ───────────────────────────────────────────────────────────

router.get('/curios/:id/digital', async (req, res) => {
  try {
    const client = getCuriosClient();
    const result = await call(client, 'GetDigitalAsset', { id: req.params.id });
    res.json(result);
  } catch (err) {
    res.status(404).json({ error: String(err) });
  }
});

// POST /curios/:id/digital/upload  — multipart/form-data with field "file"
// Uploads content to lcpserver, then registers the asset in curios-manager.
router.post('/curios/:id/digital/upload', requireAuth, upload.single('file'), async (req, res) => {
  try {
    if (!req.file) {
      res.status(400).json({ error: 'file field is required' });
      return;
    }

    const lcpServerURL = process.env.LCP_SERVER_URL;
    if (!lcpServerURL) {
      res.status(503).json({ error: 'LCP server not configured' });
      return;
    }

    // Generate a stable content UUID from the file checksum
    const checksum = createHash('sha256').update(req.file.buffer).digest('hex');
    const contentId = [
      checksum.slice(0, 8), checksum.slice(8, 12),
      checksum.slice(12, 16), checksum.slice(16, 20),
      checksum.slice(20, 32),
    ].join('-');

    // Register content with lcpserver
    const lcpRes = await fetch(`${lcpServerURL}/content/${contentId}`, {
      method: 'PUT',
      headers: { 'Content-Type': req.file.mimetype || 'application/octet-stream' },
      body: new Uint8Array(req.file.buffer),
    });
    if (!lcpRes.ok) {
      const msg = await lcpRes.text();
      res.status(502).json({ error: `lcpserver error: ${msg}` });
      return;
    }

    // Determine format from mimetype or original filename
    const ext = (req.file.originalname ?? '').split('.').pop()?.toUpperCase() ?? '';
    const format = ext || (req.file.mimetype.split('/').pop()?.toUpperCase() ?? 'BINARY');

    // Upsert the digital asset record in curios-manager
    const client = getCuriosClient();
    const asset = await call(client, 'CreateDigitalAsset', {
      curio_id: req.params.id,
      format,
      file_ref: req.file.originalname ?? '',
      checksum,
      max_concurrent: Number(req.body.maxConcurrent ?? 0),
      lcp_content_id: contentId,
      storage_backend: 'local',
      encrypted: false, // lcpserver handles encryption internally
    });

    res.status(201).json(asset);
  } catch (err) {
    res.status(500).json({ error: String(err) });
  }
});

// ─── Digital leases ───────────────────────────────────────────────────────────

router.post('/curios/:id/lease', requireAuth, async (req, res) => {
  try {
    const client = getCuriosClient();
    const result = await call(client, 'IssueLease', {
      curio_id: req.params.id,
      user_id: req.session.userId,
      user_node_id: req.body.userNodeId ?? '',
      expires_at: req.body.expiresAt ?? 0,
    });
    const lease = result as Record<string, unknown>;
    // Add computed licenseUrl if access_token is present but licenseUrl isn't set by the service
    const lsdPublicURL = process.env.LSD_PUBLIC_URL;
    if (lsdPublicURL && lease['access_token'] && !lease['license_url']) {
      lease['license_url'] = `${lsdPublicURL}/licenses/${lease['access_token']}/status`;
    }
    res.json(lease);
  } catch (err) {
    res.status(400).json({ error: String(err) });
  }
});

export default router;
