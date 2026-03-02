import express from 'express';
import cookieParser from 'cookie-parser';
import { sessionMiddleware } from './auth/session';
import authRouter from './routes/auth';
import curiosRouter from './routes/curios';
import searchRouter from './routes/search';
import peersRouter from './routes/peers';
import claimsRouter from './routes/claims';
import transfersRouter from './routes/transfers';

const app = express();
const port = process.env.PORT ?? '3001';

app.use(express.json());
app.use(cookieParser());
app.use(sessionMiddleware());

// Health check
app.get('/health', (_req, res) => res.json({ ok: true }));

// Node info — exposes public node ID so the frontend can check manager claims
app.get('/node-info', (_req, res) => {
  res.json({ nodeId: process.env.NODE_ID ?? '' });
});

app.use(authRouter);
app.use(curiosRouter);
app.use(searchRouter);
app.use(peersRouter);
app.use(claimsRouter);
app.use(transfersRouter);

app.listen(port, () => {
  console.log(`tiny-ils BFF listening on :${port}`);
});
