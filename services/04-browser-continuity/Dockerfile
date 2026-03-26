FROM node:20-alpine AS deps

WORKDIR /app
COPY package.json ./
RUN npm install --frozen-lockfile 2>/dev/null || npm install

# ---------------------------------------------------------------------------
FROM deps AS builder

WORKDIR /app
COPY tsconfig.json ./
COPY src/ ./src/
RUN npm run build

# ---------------------------------------------------------------------------
FROM node:20-alpine AS runtime

ENV NODE_ENV=production
ENV PORT=8080

WORKDIR /app

COPY package.json ./
RUN npm install --omit=dev

COPY --from=builder /app/dist ./dist

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget -qO- http://localhost:${PORT}/health || exit 1

CMD ["node", "dist/server.js"]
