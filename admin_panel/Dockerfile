FROM node:25-alpine AS base

WORKDIR /app

COPY admin_panel/package*.json ./

RUN npm install

COPY admin_panel/ .

FROM base AS development

EXPOSE 5173

CMD ["npm", "run", "dev"]

FROM base AS builder

ARG VITE_API_URL
ARG VITE_MINIO_URL
ARG VITE_GRAFANA_URL

ENV VITE_API_URL=$VITE_API_URL
ENV VITE_MINIO_URL=$VITE_MINIO_URL
ENV VITE_GRAFANA_URL=$VITE_GRAFANA_URL

RUN npm run build

FROM nginx:alpine AS production

COPY --from=builder /app/dist /usr/share/nginx/html

COPY deployment/nginx/dev.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
