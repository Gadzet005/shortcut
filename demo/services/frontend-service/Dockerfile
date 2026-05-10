FROM node:22-alpine AS builder
WORKDIR /app
COPY pkg/shortcutclient-ts ./pkg/shortcutclient-ts
COPY demo/services/frontend-service/package.json demo/services/frontend-service/tsconfig.json ./demo/services/frontend-service/
WORKDIR /app/pkg/shortcutclient-ts
RUN npm install --no-audit --no-fund && npm run build
WORKDIR /app/demo/services/frontend-service
RUN npm install --install-links --no-audit --no-fund
COPY demo/services/frontend-service/src ./src
RUN npm run build

FROM node:22-alpine
WORKDIR /app
COPY --from=builder /app/demo/services/frontend-service/node_modules ./node_modules
COPY --from=builder /app/demo/services/frontend-service/dist ./dist
COPY demo/services/frontend-service/configs ./configs
COPY demo/services/frontend-service/package.json ./package.json

EXPOSE 9104
ENV CONFIGS_DIR=./configs
CMD ["node", "dist/main.js"]
