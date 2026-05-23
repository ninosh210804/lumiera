# Архитектура ERP для кофейни — Этап 1

Этот раздел содержит архитектурные артефакты, требуемые на этапе 1
(`project.MD § Порядок работы → Этап 1`). Кода здесь нет — только схемы,
диаграммы, обоснование выбора стека и структура проекта.

## Содержание

1. [01-components.md](01-components.md) — Диаграмма компонентов системы
   (Tauri / Capacitor / Web / Go-сервер / PostgreSQL / SQLite / Event bus).
2. [02-er-diagram.md](02-er-diagram.md) — Полная ER-диаграмма базы данных
   в Mermaid с разбивкой по модулям (организация, меню, склад, рецепты,
   касса, CRM, смены, бухгалтерия, синхронизация, аудит).
3. [03-sync-engine.md](03-sync-engine.md) — Go sync engine: модель событий,
   протокол push/pull, идемпотентность, стратегии разрешения конфликтов.
4. [04-tech-stack.md](04-tech-stack.md) — Финальный стек с обоснованием
   каждого выбора (почему Go, почему sqlc, почему Tauri, почему Capacitor, …).
5. [05-project-structure.md](05-project-structure.md) — Структура монорепо
   (pnpm + Turborepo) и Go-проекта (golang-standards/project-layout).

## Чек-лист Этапа 1

- [x] Диаграмма компонентов системы → [01-components.md](01-components.md)
- [x] Полная ER-диаграмма в Mermaid → [02-er-diagram.md](02-er-diagram.md)
- [x] Описание Go sync engine и стратегии разрешения конфликтов
      → [03-sync-engine.md](03-sync-engine.md)
- [x] Финальный стек с обоснованием каждого выбора
      → [04-tech-stack.md](04-tech-stack.md)
- [x] Структура монорепо и Go-проекта → [05-project-structure.md](05-project-structure.md)

## Открытые вопросы к владельцу

См. секцию "Открытые вопросы" в конце каждого документа и сводный список
в `01-components.md` → § Открытые вопросы.
