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
├── cmd
│   ├── agent                      # основной код запуска агента
│   │   ├── flags.go               # обработка флагов командной строки для агента
│   │   ├── flags_test.go          # тесты для flags.go
│   │   ├── main.go                # точка входа агента
│   │   ├── main_test.go           # тесты для main.go
│   │   ├── run.go                 # запуск и управление жизненным циклом агента
│   │   └── run_test.go            # тесты для run.go
│   └── server                     # основной код запуска сервера
│       ├── flags.go               # обработка флагов командной строки для сервера
│       ├── flags_test.go          # тесты для flags.go сервера
│       ├── main.go                # точка входа сервера
│       ├── main_test.go           # тесты для main.go сервера
│       ├── run.go                 # запуск и управление жизненным циклом сервера
│       ├── run_test.go            # тесты для run.go сервера
├── data
│   └── metrics.json              # файл хранения метрик (json)
├── go.mod                       # управление зависимостями Go-модуля
├── go.sum                       # контрольные суммы зависимостей
├── internal                     # внутренняя реализация приложения (пакеты не для внешнего использования)
│   ├── apps                     # запускаемые приложения (агент, сервер)
│   │   ├── agent.go             # логика агента
│   │   ├── agent_test.go        # тесты агента
│   │   ├── server.go            # логика сервера
│   │   └── server_test.go       # тесты сервера
│   ├── configs                  # конфигурации для агента и сервера
│   │   ├── agent.go             # конфигурация агента
│   │   ├── server.go            # конфигурация сервера
│   │   └── server_test.go       # тесты конфигурации сервера
│   ├── contexts                 # контексты (например, транзакции)
│   │   ├── tx.go                # контекст транзакции
│   │   └── tx_test.go           # тесты для контекста транзакции
│   ├── facades                  # фасады для абстракции внешних взаимодействий
│   │   ├── metric_update.go     # фасад обновления метрик
│   │   └── metric_update_test.go # тесты фасада
│   ├── handlers                 # HTTP-обработчики и их тесты/моки
│   │   ├── metric_get_body.go   # обработчик получения метрик из тела запроса
│   │   ├── metric_get_body_mock.go # мок для тестирования
│   │   ├── metric_get_body_test.go # тесты обработчика
│   │   ├── metric_get_path.go   # обработчик получения метрик из URL пути
│   │   ├── metric_get_path_mock.go
│   │   ├── metric_get_path_test.go
│   │   ├── metric_list.go       # обработчик списка метрик
│   │   ├── metric_list_mock.go
│   │   ├── metric_list_test.go
│   │   ├── metric_update_body.go  # обработчик обновления метрик из тела
│   │   ├── metric_update_body_mock.go
│   │   ├── metric_update_body_test.go
│   │   ├── metric_update_path.go  # обработчик обновления метрик через путь
│   │   ├── metric_update_path_mock.go
│   │   ├── metric_update_path_test.go
│   │   ├── metric_updates_body.go # обработчик массового обновления метрик
│   │   ├── metric_updates_body_mock.go
│   │   ├── metric_updates_body_test.go
│   │   └── ping_db.go           # health check базы данных
│   ├── logger                   # логирование приложения
│   │   ├── logger.go            # инициализация и обертка логгера
│   │   └── logger_test.go       # тесты логгера
│   ├── middlewares              # HTTP middleware для gzip, логирования, retry и транзакций
│   │   ├── gzip.go
│   │   ├── gzip_test.go
│   │   ├── logging.go
│   │   ├── logging_test.go
│   │   ├── retry.go
│   │   ├── retry_test.go
│   │   ├── tx.go
│   │   └── tx_test.go
│   ├── repositories            # слой доступа к данным (база, файлы, память)
│   │   ├── db_helpers.go       # вспомогательные функции работы с БД
│   │   ├── db_helpers_test.go
│   │   ├── metric_db_get.go    # получение метрик из БД
│   │   ├── metric_db_get_test.go
│   │   ├── metric_db_list.go   # список метрик из БД
│   │   ├── metric_db_list_test.go
│   │   ├── metric_db_save.go   # сохранение метрик в БД
│   │   ├── metric_db_save_test.go
│   │   ├── metric_file_get.go  # получение метрик из файла
│   │   ├── metric_file_get_test.go
│   │   ├── metric_file_list.go
│   │   ├── metric_file_list_test.go
│   │   ├── metric_file_save.go
│   │   ├── metric_file_save_test.go
│   │   ├── metric_get.go       # интерфейсы и общая логика получения метрик
│   │   ├── metric_get_mock.go
│   │   ├── metric_get_test.go
│   │   ├── metric_list.go      # интерфейсы и логика списка метрик
│   │   ├── metric_list_mock.go
│   │   ├── metric_list_test.go
│   │   ├── metric_memory_get.go  # получение метрик из памяти
│   │   ├── metric_memory_get_test.go
│   │   ├── metric_memory_list.go
│   │   ├── metric_memory_list_test.go
│   │   ├── metric_memory_save.go
│   │   ├── metric_memory_save_test.go
│   │   ├── metric_save.go       # интерфейсы и логика сохранения метрик
│   │   ├── metric_save_mock.go
│   │   └── metric_save_test.go
│   ├── services                # бизнес-логика приложения (метрики)
│   │   ├── metric_get.go
│   │   ├── metric_get_mock.go
│   │   ├── metric_get_test.go
│   │   ├── metric_list.go
│   │   ├── metric_list_mock.go
│   │   ├── metric_list_test.go
│   │   ├── metric_update.go
│   │   ├── metric_update_mock.go
│   │   └── metric_update_test.go
│   ├── types                   # типы данных и структуры для метрик
│   │   ├── metrics.go
│   │   └── metrics_test.go
│   ├── validators              # валидация входных данных метрик
│   │   ├── metric.go
│   │   └── metric_test.go
│   └── workers                 # фоновые задачи агента и сервера по работе с метриками
│       ├── metric_agent.go
│       ├── metric_agent_mock.go
│       ├── metric_agent_test.go
│       ├── metric_server.go
│       ├── metric_server_mock.go
│       └── metric_server_test.go
├── Makefile                   # скрипты для сборки, тестирования, запуска и пр.
├── migrations                 # миграции базы данных
│   └── 20250612025220_create_metrics_table.sql
└── README.md                  # описание проекта
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
| iter8    | Добавление gzip middleware для cжатия/разжатия запросов/ответов | https://github.com/sbilibin2017/yp-metrics/pull/8|
| iter9    | Добавление загрузки и сохранения метрик сервером из файла | https://github.com/sbilibin2017/yp-metrics/pull/9|
| iter10    | Добавление подключения к бд (postgres) | https://github.com/sbilibin2017/yp-metrics/pull/10|
| iter11    | Добавление работы сервера с бд и стратегий(использован паттерн стратегия) хранения метрик(бд>файл>память) | https://github.com/sbilibin2017/yp-metrics/pull/11|
| iter12    | Добавление обработчика для массового обновления метрик | https://github.com/sbilibin2017/yp-metrics/pull/12|
| iter13    | Добавление обработки retriable ошибок сервера и агента | https://github.com/sbilibin2017/yp-metrics/pull/13|

---

## Развертывание проекта

1. Клонируйте репозиторий ```git clone git@github.com:sbilibin2017/yp-metrics.git```
2. Скомпилируйте сервер: ```go build -o ./cmd/server/server ./cmd/server/```
3. Скомпилируйте агент: ```go build -o ./cmd/agent/agent ./cmd/agent/```
4. Запустите сервер: ```./cmd/server/server```
5. Запустите агент: ```./cmd/agent/agent```