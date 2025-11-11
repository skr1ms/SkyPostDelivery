# Admin Panel

## Overview (EN)
- Stack: React 18, TypeScript, Vite, Tailwind-style custom CSS.
- Role: operator dashboard for monitoring drones, deliveries, parcel automats, goods, and system health.
- API: consumes orchestrator REST endpoints (`/api/v1`), Grafana panels, MinIO links.

### Useful Commands
- `npm install` — install dependencies.
- `npm run dev` — start Vite dev server on http://localhost:5173.
- `npm run build` — production build (used by Dockerfile `builder` stage).
- `npm run lint` — TS type check (used in CI).

### Key Directories
- `src/pages/` — top-level screens (Dashboard, Drones, Goods, Monitoring, etc.).
- `src/components/` — reusable UI components (Button, ConfirmModal, Layout).
- `src/api/` — Axios clients.
- `src/context/` — authentication context & hooks.
- `src/config/api_config.ts` — runtime URLs (env-driven).

### Documentation
- EN: [`docs/en/STRUCTURE.md`](../docs/en/STRUCTURE.md), [`docs/en/WORKFLOW.md`](../docs/en/WORKFLOW.md), [`docs/en/TECHNOLOGY-STACK.md`](../docs/en/TECHNOLOGY-STACK.md)
- RU: [`docs/ru/СТРУКТУРА.md`](../docs/ru/СТРУКТУРА%20ПРОЕКТА.md), [`docs/ru/СЦЕНАРИЙ РАБОТЫ.md`](../docs/ru/СЦЕНАРИЙ%20РАБОТЫ.md), [`docs/ru/ИСПОЛЬЗУЕМЫЕ ТЕХНОЛОГИИ.md`](../docs/ru/ИСПОЛЬЗУЕМЫЕ%20ТЕХНОЛОГИИ.md)

---

## Обзор (RU)
- Стек: React 18, TypeScript, Vite, кастомные CSS-компоненты.
- Роль: панель оператора для мониторинга дронов, доставок, постаматов, товаров и состояния системы.
- API: обращается к REST оркестратора (`/api/v1`), использует ссылки на Grafana и MinIO.

### Полезные команды
- `npm install` — установка зависимостей.
- `npm run dev` — дев-сервер Vite (http://localhost:5173).
- `npm run build` — продакшн-сборка (используется в Docker).
- `npm run lint` — проверка TypeScript (используется в CI).

### Основные директории
- `src/pages/` — страницы (Dashboard, Drones, Goods, Monitoring и др.).
- `src/components/` — переиспользуемые компоненты (Button, Modal, Layout).
- `src/api/` — HTTP-клиенты.
- `src/context/` — контекст авторизации.
- `src/config/api_config.ts` — конфигурация URL (ENV).

### Документация
- EN: [`docs/en/STRUCTURE.md`](../docs/en/STRUCTURE.md), [`docs/en/WORKFLOW.md`](../docs/en/WORKFLOW.md), [`docs/en/TECHNOLOGY-STACK.md`](../docs/en/TECHNOLOGY-STACK.md)
- RU: [`docs/ru/СТРУКТУРА.md`](../docs/ru/СТРУКТУРА%20ПРОЕКТА.md), [`docs/ru/СЦЕНАРИЙ РАБОТЫ.md`](../docs/ru/СЦЕНАРИЙ%20РАБОТЫ.md), [`docs/ru/ИСПОЛЬЗУЕМЫЕ ТЕХНОЛОГИИ.md`](../docs/ru/ИСПОЛЬЗУЕМЫЕ%20ТЕХНОЛОГИИ.md)
