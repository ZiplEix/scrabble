FROM oven/bun:latest AS builder
WORKDIR /app
COPY package.json bun.lock ./
RUN bun install --frozen-lockfile
COPY . .

ARG PUBLIC_API_BASE_URL
ARG PUBLIC_VAPID_PUBLIC_KEY
ARG VITE_API_BASE_URL
ENV PUBLIC_API_BASE_URL=$PUBLIC_API_BASE_URL
ENV PUBLIC_VAPID_PUBLIC_KEY=$PUBLIC_VAPID_PUBLIC_KEY
ENV VITE_API_BASE_URL=$VITE_API_BASE_URL

RUN bun run build

FROM oven/bun:latest
WORKDIR /app
COPY --from=builder /app/build build/
COPY --from=builder /app/node_modules node_modules/
COPY package.json .
EXPOSE 3000
ENV NODE_ENV=production
ENV PORT=3000
CMD [ "bun", "run", "build/index.js" ]
