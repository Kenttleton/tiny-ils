FROM node:22-alpine AS builder
WORKDIR /app
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/svelte.config.js ./
COPY frontend/vite.config.ts ./
COPY frontend/tsconfig.json ./
COPY frontend/src/ ./src/
COPY frontend/static/ ./static/
RUN npm run build

FROM node:22-alpine
WORKDIR /app
COPY frontend/package*.json ./
RUN npm ci --omit=dev
COPY --from=builder /app/build ./build
COPY proto/ ./proto/
EXPOSE 3000
ENV HOST=0.0.0.0 PORT=3000
CMD ["node", "build"]
