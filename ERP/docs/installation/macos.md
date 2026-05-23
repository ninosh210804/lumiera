# Установка Coffeeshop ERP на macOS

## Требования

| Компонент | Минимум |
|-----------|---------|
| ОС | macOS 12 Monterey или новее |
| Чип | Intel x64 или Apple Silicon (M1/M2/M3) |
| RAM | 4 GB |
| Диск | 500 MB |

---

## Вариант 1 — Установка с флешки

### Шаг 1. Скопируйте .dmg

```
coffeeshop-erp-v1.0.0-macos.dmg         # Intel
coffeeshop-erp-v1.0.0-macos-arm64.dmg   # Apple Silicon (M1/M2/M3)
```

Проверьте чип: **Apple → Об этом Mac** → поле «Процессор» или «Чип».

### Шаг 2. Откройте DMG

Дважды щёлкните по файлу `.dmg`.  
Перетащите **Coffeeshop ERP** в папку **Applications** (Приложения).

### Шаг 3. Обход Gatekeeper

macOS заблокирует первый запуск непроверенного приложения:

> **«Coffeeshop ERP» невозможно открыть, так как Apple не может проверить его»**

**Способ А** (рекомендуется):
1. Найдите приложение в папке **Applications**
2. Удерживайте **Control** и щёлкните по значку
3. Выберите **«Открыть»**
4. В диалоге нажмите **«Открыть»**

**Способ Б** (через Терминал):
```bash
xattr -d com.apple.quarantine /Applications/Coffeeshop\ ERP.app
```

### Шаг 4. Первый запуск

Приложение запустится. Значок появится в строке меню (вверху справа).

---

## Вариант 2 — Запуск через Терминал (без установки)

```bash
# Для Intel
./erp-server-darwin-amd64 --port 8080

# Для Apple Silicon
./erp-server-darwin-arm64 --port 8080
```

Откройте браузер: `http://localhost:8080`

---

## Настройка базы данных

```bash
# PostgreSQL (рекомендуется для сервера)
export DATABASE_URL="postgres://postgres:пароль@localhost:5432/coffeeshop"
./erp-server --migrate

# SQLite (для одной кассы без сервера)
./erp-server --sqlite --sqlite-path ~/erp-data/coffeeshop.db
```

---

## Автозапуск при входе в систему

```bash
# Создать LaunchAgent
cat > ~/Library/LaunchAgents/kz.coffeeshop.erp.plist << 'EOF'
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key><string>kz.coffeeshop.erp</string>
  <key>ProgramArguments</key>
  <array>
    <string>/Applications/Coffeeshop ERP.app/Contents/MacOS/erp-server</string>
  </array>
  <key>RunAtLoad</key><true/>
  <key>KeepAlive</key><true/>
</dict>
</plist>
EOF

launchctl load ~/Library/LaunchAgents/kz.coffeeshop.erp.plist
```

---

## Удаление

```bash
# Удалить приложение
rm -rf /Applications/Coffeeshop\ ERP.app

# Удалить LaunchAgent (если настроен)
launchctl unload ~/Library/LaunchAgents/kz.coffeeshop.erp.plist
rm ~/Library/LaunchAgents/kz.coffeeshop.erp.plist

# Данные SQLite (если использовался)
rm -rf ~/erp-data/
```

---

## Решение проблем

| Проблема | Решение |
|----------|---------|
| «Приложение повреждено» | Запустите: `xattr -cr /Applications/Coffeeshop\ ERP.app` |
| Порт 8080 занят | `lsof -i :8080` → `kill -9 <PID>`, затем перезапустите |
| Нет прав на запись | `chmod +x erp-server-darwin-*` |
| WebKit не загружает UI | Откройте Safari → Настройки → Конфиденциальность → разрешите localhost |

---

*Версия документа: 1.0 | Поддержка: admin@coffeeshop.kz*
