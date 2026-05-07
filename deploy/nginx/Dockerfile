FROM node:20.18-alpine AS frontend-builder

WORKDIR /app

COPY frontend/package.json frontend/package-lock.json* ./
RUN npm ci

COPY frontend ./
RUN npm run build

FROM nginx:1.27-alpine

ARG NGINX_CONF=nginx.conf

COPY deploy/nginx/${NGINX_CONF} /etc/nginx/templates/default.conf.template
COPY --from=frontend-builder /app/dist /usr/share/nginx/html
