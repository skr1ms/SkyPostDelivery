export const API_CONFIG = {
  baseURL: import.meta.env.VITE_API_URL || '/api/v1',
  timeout: 10000,
};

export const MONITORING_CONFIG = {
  minioURL: import.meta.env.VITE_MINIO_URL || 'http://localhost:9001',
  grafanaURL: import.meta.env.VITE_GRAFANA_URL || 'http://localhost:3001',
  rabbitmqURL: import.meta.env.VITE_RABBITMQ_URL || 'http://localhost:15672',
};

export const APP_CONFIG = {
  name: 'SkyPost Delivery Admin Panel',
  version: '1.0.0',
};
