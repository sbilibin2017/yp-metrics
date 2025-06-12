# Яндекс Практикум "Продвинутый Go Разработчик"  
## Сервис сбора метрик и алертинга

## 📌 Описание проекта

Это система для сбора runtime-метрик, разработанная в рамках трека **«Сервис сбора метрик и алертинга»** курса **«Продвинутый Go-разработчик»** от Яндекс Практикума.  
Состоит из двух компонентов:

- **Сервер** — принимает и хранит метрики;
- **Агент** — собирает метрики из среды выполнения и отправляет их на сервер.

---

## ⚙️ Метрики

Поддерживаются два типа метрик:

- **gauge** (`float64`) — значение перезаписывается при каждом обновлении;
- **counter** (`int64`) — значение увеличивается на заданное при каждом обновлении.

---

## 🌐 Сервер

Сервер слушает `http://localhost:8080` и обрабатывает http запросы  

---

## 🛰 Агент

Агент предназначен для автоматического сбора метрик из стандартной библиотеки `runtime` и их периодической отправки на сервер.

### Что делает агент:

- Использует `runtime.ReadMemStats` для сбора метрик (`Alloc`, `TotalAlloc`, `Sys`, и др.);
- Обновляет значения `counter` метрик (например, количество попыток отправки);
- Отправляет данные на сервер с заданной периодичностью (`pollInterval`, `reportInterval`);
- Работает параллельно, используя `context.Context` и фоновые воркеры.

Агент можно запустить отдельно, указав адрес сервера и частоту опроса/отправки метрик через конфигурацию.

---

## Структура проекта

Проект имеет слоистую архитектуру, взаимодействие слоев осуществляется с помощью интерфейсов


```
.
├── cmd                                # Точка входа в приложение (agent и server)
│   ├── agent                          # Каталог с исходниками агента
│   │   ├── agent                      # Логика агента (сбор метрик, отправка)
│   │   ├── main.go                    # Главный файл запуска агента
│   │   └── main_test.go              # Тесты для запуска агента
│   └── server                         # Каталог с исходниками сервера
│       ├── main.go                    # Главный файл запуска сервера
│       ├── main_test.go              # Тесты для запуска сервера
│       └── server                    # Логика запуска HTTP-сервера (может быть роутинг, инициализация)
├── data
│   └── metrics.json                   # JSON-файл для хранения метрик при использовании файлового стора
├── go.mod                             # Go-модуль (описание зависимостей)
├── go.sum                             # Контрольные суммы для зависимостей
├── internal                           # Внутренние пакеты (не экспортируются за пределы модуля)
│   ├── contexts
│   │   ├── tx.go                      # Контекст для работы с транзакциями
│   │   └── tx_test.go                # Тесты для контекста транзакций
│   ├── facades
│   │   ├── metric_update.go          # Фасад (обертка) для обновления метрик
│   │   └── metric_update_test.go     # Тесты фасада обновления метрик
│   ├── handlers                       # HTTP-обработчики
│   │   ├── metric_get_body.go        # POST /value/ - получение метрик по JSON
│   │   ├── metric_get_body_mock.go   # Моки для тестов metric_get_body
│   │   ├── metric_get_body_test.go   # Тесты для metric_get_body
│   │   ├── metric_get_path.go        # GET /value/{type}/{name} - получение метрик по пути
│   │   ├── metric_get_path_mock.go   # Моки для metric_get_path
│   │   ├── metric_get_path_test.go   # Тесты для metric_get_path
│   │   ├── metric_list.go            # GET / - HTML-список всех метрик
│   │   ├── metric_list_mock.go       # Моки для списка метрик
│   │   ├── metric_list_test.go       # Тесты списка метрик
│   │   ├── metric_update_body.go     # POST /update/ - обновление метрики по JSON
│   │   ├── metric_update_body_mock.go# Моки для metric_update_body
│   │   ├── metric_update_body_test.go# Тесты для metric_update_body
│   │   ├── metric_update_path.go     # POST /update/{type}/{name}/{value} - обновление по пути
│   │   ├── metric_update_path_mock.go# Моки для metric_update_path
│   │   └── metric_update_path_test.go# Тесты для metric_update_path
│   ├── logger
│   │   ├── logger.go                 # Настройка логгера (уровень, формат и т.п.)
│   │   └── logger_test.go           # Тесты логгера
│   ├── middlewares
│   │   ├── gzip.go                  # Middleware для сжатия/распаковки gzip
│   │   ├── gzip_test.go            # Тесты gzip middleware
│   │   ├── logging.go              # Middleware логирования HTTP-запросов
│   │   ├── logging_test.go         # Тесты логирования
│   │   ├── tx.go                   # Middleware для управления транзакциями
│   │   └── tx_test.go              # Тесты транзакционного middleware
│   ├── repositories                 # Реализация хранилищ метрик (память, файл, БД)
│   │   ├── db_helpers.go           # Вспомогательные функции для работы с БД
│   │   ├── db_helpers_test.go      # Тесты для db_helpers
│   │   ├── metric_db_get.go        # Получение метрик из БД
│   │   ├── metric_db_get_test.go   # Тесты metric_db_get
│   │   ├── metric_db_list.go       # Получение списка метрик из БД
│   │   ├── metric_db_list_test.go  # Тесты metric_db_list
│   │   ├── metric_db_save.go       # Сохранение метрик в БД
│   │   ├── metric_db_save_test.go  # Тесты metric_db_save
│   │   ├── metric_file_get.go      # Получение метрик из файла
│   │   ├── metric_file_get_test.go # Тесты metric_file_get
│   │   ├── metric_file_list.go     # Список метрик из файла
│   │   ├── metric_file_list_test.go# Тесты metric_file_list
│   │   ├── metric_file_save.go     # Сохранение метрик в файл
│   │   ├── metric_file_save_test.go# Тесты metric_file_save
│   │   ├── metric_get.go           # Общее получение метрик (интерфейсы)
│   │   ├── metric_get_mock.go      # Моки
│   │   ├── metric_get_test.go      # Тесты
│   │   ├── metric_list.go          # Получение списка метрик
│   │   ├── metric_list_mock.go     # Моки
│   │   ├── metric_list_test.go     # Тесты
│   │   ├── metric_memory_get.go    # Получение из памяти
│   │   ├── metric_memory_get_test.go# Тесты
│   │   ├── metric_memory_list.go   # Список из памяти
│   │   ├── metric_memory_list_test.go# Тесты
│   │   ├── metric_memory_save.go   # Сохранение в память
│   │   ├── metric_memory_save_test.go# Тесты
│   │   ├── metric_save.go          # Универсальный интерфейс сохранения
│   │   ├── metric_save_mock.go     # Моки
│   │   └── metric_save_test.go     # Тесты
│   ├── services
│   │   ├── metric_get.go           # Сервис получения метрик (агрегация логики)
│   │   ├── metric_get_mock.go      # Моки
│   │   ├── metric_get_test.go      # Тесты
│   │   ├── metric_list.go          # Сервис списка метрик
│   │   ├── metric_list_mock.go     # Моки
│   │   ├── metric_list_test.go     # Тесты
│   │   ├── metric_update.go        # Сервис обновления метрик
│   │   ├── metric_update_mock.go   # Моки
│   │   └── metric_update_test.go   # Тесты
│   ├── types
│   │   ├── metrics.go              # Описание структур метрик
│   │   └── metrics_test.go         # Тесты для структур
│   ├── validators
│   │   ├── metric.go               # Валидация метрик (тип, имя, значение)
│   │   └── metric_test.go          # Тесты валидатора
│   └── workers
│       ├── metric_agent.go         # Работа агентской стороны (сбор метрик)
│       ├── metric_agent_mock.go    # Моки
│       ├── metric_agent_test.go    # Тесты
│       ├── metric_server.go        # Обработка на стороне сервера
│       ├── metric_server_mock.go   # Моки
│       └── metric_server_test.go   # Тесты
├── Makefile                         # Сценарии сборки и запуска
├── migrations
│   └── 20250612025220_create_metrics_table.sql # SQL-миграция для БД
└── README.md                        # Документация по проекту
                  # Документация проекта
```

---

## Используемые технологии

| Технология / Библиотека | Описание                              |
|-------------------------|-------------------------------------|
| Go                      | Язык программирования                |
| Chi                     | HTTP роутер                         |
| Resty                   | HTTP клиент                         |
| Testify                 | Фреймворк для тестирования          |
| Docker                  | Утилита для контейнеризации         |

---

## Инкременты проекта

| Итерация | Описание                                               | Ссылка на PR                         |
|----------|--------------------------------------------------------|------------------------------------|
| iter1    | Реализация сервера сбора метрик с HTTP API             | https://github.com/sbilibin2017/yp-metrics/pull/1|
| iter2    | Реализация агента сбора метрик(использован паттерн фасад) | https://github.com/sbilibin2017/yp-metrics/pull/2|
| iter3    | Добавление к серверу обработчиков дял поулчения метрик | https://github.com/sbilibin2017/yp-metrics/pull/3|
| iter4    | Добавление флагов для конфигурирования серва и агнта   | https://github.com/sbilibin2017/yp-metrics/pull/4|
| iter5    | Добавление приоритета конфигураций (env>flag>default)  | https://github.com/sbilibin2017/yp-metrics/pull/5|
| iter6    | Добавление logging middleware для логирования запросов и ответов сервера  | https://github.com/sbilibin2017/yp-metrics/pull/6|
| iter7    | Добавление обновление и получение метрик в теле запроса | https://github.com/sbilibin2017/yp-metrics/pull/7|
| iter8    | Добавление gzip middleware для зжатия/разжатия запросов/ответов | https://github.com/sbilibin2017/yp-metrics/pull/8|
| iter9    | Добавление загрузки и сохранения метрик сервером из файла | https://github.com/sbilibin2017/yp-metrics/pull/9|
| iter10    | Добавление подключения к бд (postgres) | https://github.com/sbilibin2017/yp-metrics/pull/10|
| iter11    | Добавление работы сервера с бд и стратегий(использован паттерн стратегия) хранения метрик(бд>файл>память) | https://github.com/sbilibin2017/yp-metrics/pull/11|

---

## Развертывание проекта

1. Клонируйте репозиторий ```git clone git@github.com:sbilibin2017/yp-metrics.git```
2. Скомпилируйте сервер: ```go build -o ./cmd/server/server ./cmd/server/```
3. Скомпилируйте агент: ```go build -o ./cmd/agent/agent ./cmd/agent/```
4. Запустите сервер: ```./cmd/server/server```
5. Запустите агент: ```./cmd/agent/agent```