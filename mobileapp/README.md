# Mobile App

## Overview (EN)
- Framework: Flutter (Dart).
- Platforms: Android, iOS, Web (single codebase).
- Role: customer/courier app for registration, browsing goods, placing orders, tracking deliveries, confirming pickup.
- Architecture: feature-based folders with Clean Architecture layers (`data`, `domain`, `presentation`).
- Networking: custom `HttpClient` with auth interceptors, integrates with orchestrator REST API and push notifications.

### Useful Commands
- `flutter pub get` — install dependencies.
- `flutter run` — launch on connected device or emulator.
- `flutter test` — run widget/unit tests.
- `flutter build apk` / `flutter build ios` / `flutter build web` — production builds.

### Key Directories
- `lib/core/` — DI container, networking, theme, push service.
- `lib/features/` — feature modules (auth, goods, orders, qr, notifications, etc.).
- `assets/` — static resources.
- `android/`, `ios/`, `web/` — platform-specific projects.
- `test/` — automated tests.

### Documentation
- EN: [`docs/en/STRUCTURE.md`](../docs/en/STRUCTURE.md), [`docs/en/WORKFLOW.md`](../docs/en/WORKFLOW.md)
- RU: [`docs/ru/СТРУКТУРА.md`](../docs/ru/СТРУКТУРА%20ПРОЕКТА.md), [`docs/ru/СЦЕНАРИЙ РАБОТЫ.md`](../docs/ru/СЦЕНАРИЙ%20РАБОТЫ.md)

---

## Обзор (RU)
- Фреймворк: Flutter (Dart).
- Платформы: Android, iOS, Web (единая кодовая база).
- Роль: приложение пользователя/курьера — регистрация, выбор товаров, оформление заказов, отслеживание доставки, подтверждение получения.
- Архитектура: фичи с разделением на слои `data`, `domain`, `presentation` (Clean Architecture).
- Сети: собственный `HttpClient` с перехватчиками авторизации, интеграция с REST оркестратора и push-уведомлениями.

### Полезные команды
- `flutter pub get` — установка зависимостей.
- `flutter run` — запуск на устройстве/эмуляторе.
- `flutter test` — виджет/юнит тесты.
- `flutter build apk` / `flutter build ios` / `flutter build web` — продакшн-сборки.

### Основные каталоги
- `lib/core/` — DI, сетевой слой, темы, push-сервис.
- `lib/features/` — модульные фичи (auth, goods, orders, qr, notifications и т.д.).
- `assets/` — статические ресурсы.
- `android/`, `ios/`, `web/` — специфичные проекты.
- `test/` — автоматические тесты.

### Документация
- EN: [`docs/en/STRUCTURE.md`](../docs/en/STRUCTURE.md), [`docs/en/WORKFLOW.md`](../docs/en/WORKFLOW.md)
- RU: [`docs/ru/СТРУКТУРА.md`](../docs/ru/СТРУКТУРА%20ПРОЕКТА.md), [`docs/ru/СЦЕНАРИЙ РАБОТЫ.md`](../docs/ru/СЦЕНАРИЙ%20РАБОТЫ.md)
