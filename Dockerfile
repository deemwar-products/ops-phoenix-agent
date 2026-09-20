# React UI — build the Vite SPA, serve the static dist with nginx.
FROM node:20-alpine AS builder
WORKDIR /app

# All Vite_* env vars are baked in at build time.
ARG VITE_API_URL
ENV VITE_API_URL=$VITE_API_URL

COPY package.json package-lock.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM nginx:alpine
RUN rm -rf /usr/share/nginx/html/*
COPY --from=builder /app/dist /usr/share/nginx/html
# Installed as a TEMPLATE, not a conf: the stock nginx entrypoint runs envsubst
# over /etc/nginx/templates/*.template into /etc/nginx/conf.d/ at container
# start, which is how ${API_UPSTREAM} becomes a deploy-time value instead of a
# hostname baked into the image.
#
# API_UPSTREAM is deliberately NOT defaulted — it is deployment-specific.
ENV DNS_RESOLVER=127.0.0.11
COPY docker-entrypoint.d/15-require-api-upstream.sh /docker-entrypoint.d/15-require-api-upstream.sh
COPY nginx.conf.template /etc/nginx/templates/default.conf.template
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
