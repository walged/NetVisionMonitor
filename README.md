<p align="center">
  <img src="logo.png" alt="NetVisionMonitor Logo" width="180" height="180">
</p>

<h1 align="center">NetVisionMonitor</h1>

<p align="center">
  <strong>🌐 Система мониторинга сетевой инфраструктуры</strong>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Platform-Windows-blue?style=for-the-badge&logo=windows" alt="Platform">
  <img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go" alt="Go Version">
  <img src="https://img.shields.io/badge/React-18+-61DAFB?style=for-the-badge&logo=react" alt="React">
  <img src="https://img.shields.io/badge/Wails-v2-red?style=for-the-badge" alt="Wails">
  <img src="https://img.shields.io/badge/License-MIT-green?style=for-the-badge" alt="License">
</p>

<p align="center">
  <a href="#-возможности">Возможности</a> •
  <a href="#-скриншоты">Скриншоты</a> •
  <a href="#-установка">Установка</a> •
  <a href="#-сборка">Сборка</a> •
  <a href="#-технологии">Технологии</a>
</p>

---

## 📖 Описание

**NetVisionMonitor** — это современное десктопное приложение для мониторинга сетевой инфраструктуры видеонаблюдения. Разработано специально для управления IP-камерами, PoE-коммутаторами, серверами и другим сетевым оборудованием.

Приложение обеспечивает непрерывный контроль состояния устройств, оповещения при сбоях, детальную статистику и удобную визуализацию сетевой топологии.

---

## ✨ Возможности

### 🔍 Мониторинг устройств
- **ICMP Ping** — проверка доступности устройств с настраиваемыми интервалами
- **SNMP v1/v2c/v3** — сбор расширенной информации с коммутаторов
- **Автоматическое обнаружение** — статус устройств в реальном времени
- **История состояний** — графики uptime и латентности

### 🖥️ Поддерживаемые устройства
| Тип | Описание |
|-----|----------|
| 📷 **Камеры** | IP-камеры видеонаблюдения |
| 🔌 **Коммутаторы** | PoE-коммутаторы с SNMP |
| 🖥️ **Серверы** | Серверы записи и хранения |

### 📊 Аналитика и статистика
- Графики латентности в реальном времени
- Статистика доступности (uptime)
- История событий с фильтрацией
- Экспорт данных

### 🗺️ Визуализация
- **Интерактивная схема** — drag & drop размещение устройств
- **Карта сети** — автоматическая топология связей
- **Связи портов** — визуализация подключений камер к коммутаторам
- **SFP Uplink** — отображение межкоммутаторных соединений

### ⚙️ SNMP мониторинг коммутаторов
```
┌─────────────────────────────────────────────────────┐
│  📊 Информация с коммутатора                        │
├─────────────────────────────────────────────────────┤
│  • Статус портов (up/down)                          │
│  • Скорость подключения (10/100/1000 Mbps)          │
│  • PoE потребление на порт                          │
│  • RX/TX трафик                                     │
│  • Температура и напряжение                         │
│  • Информация о системе                             │
└─────────────────────────────────────────────────────┘
```

### 🔔 Уведомления
- Звуковые оповещения при сбоях
- Всплывающие уведомления
- Иконка в системном трее с индикацией статуса
- Журнал событий

### 🎨 Интерфейс
- Современный тёмный дизайн
- Адаптивный интерфейс
- Поддержка светлой/тёмной темы
- Сворачивание в системный трей

---

## 📸 Скриншоты

<details>
<summary><b>🖼️ Нажмите для просмотра скриншотов</b></summary>

### Главная страница — Устройства
![Devices](docs/screenshots/devices.png)

### Детали устройства
![Device Details](docs/screenshots/device-details.png)

### Схема сети
![Schema](docs/screenshots/schema.png)

### Настройки
![Settings](docs/screenshots/settings.png)

</details>

---

## 🚀 Установка

### Готовый релиз

1. Скачайте последний релиз из [Releases](https://github.com/yourusername/NetVisionMonitor/releases)
2. Распакуйте архив
3. Запустите `NetVisionMonitor.exe`

### Portable режим

Приложение автоматически определяет portable режим, если рядом с exe-файлом находится папка `data/`.

---

## 🔧 Сборка из исходников

### Требования

- **Go** 1.21+
- **Node.js** 18+
- **Wails CLI** v2

### Установка Wails

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

### Клонирование и сборка

```bash
# Клонировать репозиторий
git clone https://github.com/yourusername/NetVisionMonitor.git
cd NetVisionMonitor

# Установить зависимости frontend
cd frontend && npm install && cd ..

# Собрать приложение
wails build
```

Готовый exe-файл будет в папке `build/bin/`.

### Режим разработки

```bash
wails dev
```

---

## 🛠️ Технологии

<table>
<tr>
<td align="center" width="150">
<img src="https://go.dev/blog/go-brand/Go-Logo/PNG/Go-Logo_Blue.png" width="60"><br>
<b>Go</b><br>
<sub>Backend</sub>
</td>
<td align="center" width="150">
<img src="https://upload.wikimedia.org/wikipedia/commons/a/a7/React-icon.svg" width="60"><br>
<b>React</b><br>
<sub>Frontend</sub>
</td>
<td align="center" width="150">
<img src="https://www.vectorlogo.zone/logos/typescriptlang/typescriptlang-icon.svg" width="60"><br>
<b>TypeScript</b><br>
<sub>Language</sub>
</td>
<td align="center" width="150">
<img src="https://wails.io/img/logo-universal.png" width="60"><br>
<b>Wails v2</b><br>
<sub>Framework</sub>
</td>
</tr>
<tr>
<td align="center" width="150">
<img src="https://tailwindcss.com/_next/static/media/tailwindcss-mark.3c5441fc7a190e4ced9600b6ed2a4417.svg" width="60"><br>
<b>Tailwind CSS</b><br>
<sub>Styling</sub>
</td>
<td align="center" width="150">
<img src="https://ui.shadcn.com/apple-touch-icon.png" width="60"><br>
<b>shadcn/ui</b><br>
<sub>Components</sub>
</td>
<td align="center" width="150">
<img src="https://www.sqlite.org/images/sqlite370_banner.gif" width="80"><br>
<b>SQLite</b><br>
<sub>Database</sub>
</td>
<td align="center" width="150">
<img src="https://gosnmp.github.io/gosnmp/gosnmp-logo.png" width="60"><br>
<b>GoSNMP</b><br>
<sub>SNMP Client</sub>
</td>
</tr>
</table>

---

## 📁 Структура проекта

```
NetVisionMonitor/
├── 📂 frontend/           # React frontend
│   ├── 📂 src/
│   │   ├── 📂 components/ # UI компоненты
│   │   ├── 📂 pages/      # Страницы приложения
│   │   ├── 📂 hooks/      # React hooks
│   │   └── 📂 lib/        # Утилиты
│   └── 📄 package.json
├── 📂 internal/           # Go internal packages
│   ├── 📂 database/       # SQLite операции
│   ├── 📂 monitoring/     # Ping & SNMP мониторинг
│   ├── 📂 models/         # Структуры данных
│   ├── 📂 snmp/           # SNMP клиент
│   └── 📂 logger/         # Логирование
├── 📂 build/              # Ресурсы сборки
│   └── 📂 windows/        # Windows иконки
├── 📄 app.go              # Основной App struct
├── 📄 main.go             # Точка входа
└── 📄 wails.json          # Конфигурация Wails
```

---

## ⚙️ Конфигурация

### Настройки мониторинга

| Параметр | Описание | По умолчанию |
|----------|----------|--------------|
| Интервал пинга | Частота проверки устройств | 30 сек |
| Таймаут | Время ожидания ответа | 5 сек |
| Попытки | Количество повторов | 3 |
| SNMP интервал | Частота опроса SNMP | 60 сек |

### SNMP v3 параметры

Поддерживаются все уровни безопасности:
- `noAuthNoPriv` — без аутентификации
- `authNoPriv` — только аутентификация (MD5/SHA)
- `authPriv` — аутентификация + шифрование (DES/AES)

---

## 🔐 Безопасность

- Учётные данные SNMP шифруются AES-256-GCM
- Ключ шифрования уникален для каждой установки
- Данные хранятся локально в SQLite

---

## 📝 Лицензия

Этот проект распространяется под лицензией MIT. Подробнее см. [LICENSE](LICENSE).

---

## 🤝 Вклад в проект

Приветствуются любые предложения и улучшения!

1. Fork репозитория
2. Создайте ветку (`git checkout -b feature/amazing-feature`)
3. Commit изменения (`git commit -m 'Add amazing feature'`)
4. Push в ветку (`git push origin feature/amazing-feature`)
5. Откройте Pull Request

---

## 📞 Контакты

Если у вас есть вопросы или предложения, создайте [Issue](https://github.com/yourusername/NetVisionMonitor/issues).

---

<p align="center">
  <sub>Made with ❤️ for network administrators</sub>
</p>
