# Data Extension Service (DES) 

**Сервис** занимаеться обогощением данных, получаемых по средствам <br> gRPC запросов.

## Как его собрать/запустить?  

### Интерактивный запуск

```bash
    # чтобы запустить файл
    go run cmd/des/des.go
```

### Запуск из исходника

```bash 
    # чтобы собрать исполняемый файл
    go build cmd/des/des.go
    # запуск исполняемого файла
    ./des
```

Также к проекту приложен `Makefile`


## ![Typing SVG](https://readme-typing-svg.herokuapp.com?color=%2336BCF7&lines=Что+реализовано+в+проекте?)

## [Config package](https://github.com/EgorKo25/DES/blob/main/internal/config/config.go)

### Описание

```go
package config

type AppConfig struct {
  WorkerConfig       WorkerConfig  `json:"worker"`
  ServiceConfig      ServiceConfig `json:"service"`
  ChannelSize        int           `json:"queue_task_size"`
  CacheClearInterval int           `json:"cache_clear_interval"`
}

type ServiceConfig struct {
  IP   string `json:"IP"`
  PORT string `json:"PORT"`
}

type WorkerConfig struct {
  RemoteHTTPServer struct {
    IP   string `json:"IP"`
    PORT string `json:"PORT"`
  } `json:"remote_http_server"`
  Authentication struct {
    Login    string `json:"login"`
    Password string `json:"password"`
  } `json:"authentication"`
  MaxWorkers         int `json:"max_workers"`
  MaxTimeForResponse int `json:"max_time_for_response"`
  TimeoutConnection  int `json:"timeout_connection"`
}
```

`AppConfig`: Основная структура, содержащая все параметры конфигурации.

`ServiceConfig`: Подструктура, описывающая параметры конфигурации сервиса.

`WorkerConfig`: Подструктура, описывающая параметры конфигурации рабочего процесса, включая параметры для удаленного HTTP-сервера и аутентификации.

### Флаг командной строки
`-path-to-config`: Флаг для указания пути к файлу конфигурации. Если флаг не указан, будет использован путь по умолчанию *(config/app.conf)*.

### Важное замечание

+ **Конфигурация загружается из файла JSON**. Убедитесь, что файл конфигурации существует и имеет правильный формат.

### Пример конфигурации

```json

{
    "channel_size":20,
    "cache_clear_interval":20,
    "service": {
        "IP": "127.0.0.1",
        "PORT": ":8080"
    },
    "worker": {
        "_comments": {
            "description": "this is a configuration for worker pull"
        },
        "max_workers": 6,
        "timeout_connection": 0,
        "max_time_for_response": 2,
        "remote_http_server": {
            "_comments": {
                "description": "this is a remote http server config"
            },
            "IP": "127.0.0.1",
            "PORT": "4557"
        },
        "authentication": {
            "_comments": {
                "description": "this is a data for authentication"
            },
            "login": "admin",
            "password": "admin"
        }
    
}
```
#### Значения полей:
+ `channel_size` - максимальный размер очереди задач
+ `cache_clear_interval` - интервал очистки кэша
+ `service` - адрес и порт, на которых работает сервис
+ `worker` - 
  + `max_workers` - максимальное количесвто параллельно работающих воркеров
  + `timeout_connection` - таймаут подключения к удаленному **HTTP**(0 - keep alive соединение)
  + `max_time_for_responce` - максимальное время для ответа воркера, если время выйдет, будет отменён контекст
  + `remote_http_server` - адресс и порт сервера при помощик которого, будут обогощаться данные
  + `authentication` - данные для аунтефикации на удаленном **HTTP** (передаються в заголовке `Authorization` в **base64**)
## [Logger package](https://github.com/EgorKo25/DES/blob/main/internal/logger/logger.go)

Этот пакет предоставляет простой и гибкий механизм логгирования на основе библиотеки [Zap](https://pkg.go.dev/go.uber.org/zap).

### Конфигурация логгера

- **Консольный вывод:**
  Логгер выводит сообщения на консоль с использованием консольного кодировщика для более читаемого вывода.

- **Файловый вывод:**
  Логгер также записывает сообщения в файлы для отслеживания деталей работы приложения. Различаются логи по уровням: `debug.log` для отладочных сообщений, `warning_error.log` для предупреждений и ошибок, а также `http.log` для отслеживания всех запросов к стороннему **HTTP** и `grpc.log` для отслеживания запросов по **grpc**.

- **Уровни логгирования:**
  Логгер настроен для разделения сообщений на уровни отладки, информации, предупреждений и ошибок.


Этот пакет предоставляет стандартный способ создания и настройки логгеров в вашем приложении, обеспечивая удобство использования и гибкость в конфигурации.
## [Cache package](https://github.com/EgorKo25/DES/blob/main/internal/cache/cache.go)

Этот пакет предоставляет простой механизм кэширования данных с автоматической очисткой по истечении указанного времени.

### Функционал
1. **Создание кэша:**
   Используйте функцию `NewCache` для создания нового экземпляра кэша:

   ```go
   c := cache.NewCache(ctx, duration)
   ```

3. **Загрузка данных в кэш:**
   Используйте метод `Load` для добавления данных в кэш:

   ```go
   Load(title, data)
   ```

4. **Поиск данных в кэше:**
   Используйте метод `Search` для поиска данных в кэше:

   ```go
   data, ok := c.Search(title)
   ```

5. **Автоматическая очистка:**
   Кэш автоматически очищается от данных по истечении указанного времени с момента последней загрузки(`duration`). Время указывается при создании кэша.

## [Service package](https://github.com/EgorKo25/DES/blob/main/internal/server/service/service.go)

Давайте разберем каждую функцию и структуру в вашем коде:

1. структура **`ExtServer`:**
```go
type ExtServer struct {
	pb.UnimplementedUserExtensionServiceServer

	logger     *zap.Logger
	grpcLogger *zap.Logger

	cache Cacher
}
```
  - `logger`: Экземпляр логгера из библиотеки `go.uber.org/zap` для логирования общих событий сервиса.

  - `grpcLogger`: Еще один экземпляр логгера для логирования событий gRPC, таких как начало и завершение соединения, а также информации о запросах.

  - `cache`: Интерфейс `Cacher`, который представляет кэш для хранения данных. Этот интерфейс предоставляет два метода: `Load` для добавления данных в кэш и `Search` для поиска данных по ключу.
2. Интерфейс **`Cacher`:**
```go
type Cacher interface {
	Load(title string, data any)
	Search(title string) (data any, ok bool)
}
```
Интерфейс призван уменьшить связность пакетов и отвязать сервис от конеретной реализации кэширования.
3. функция **`NewExtServer`:**

```go
func NewExtServer(channel chan chan []byte, logger, grpcLogger *zap.Logger, cache Cacher) *ExtServer
```

  - Создает новый экземпляр `ExtServer`.

  - Принимает канал `channel` для взаимодействия с воркерами и логгеры для логирования.

  - Возвращает указатель на новый экземпляр `ExtServer`.

4. метод-обработчик **`GetUserExtension`:**
```go
func (es *ExtServer) GetUserExtension(ctx context.Context, in *pb.GetRequest) (out *pb.GetResponse, err error)
```
  - Метод, удовлетворяющий интерфейсу `UserExtensionServiceServer` сгенерированного gRPC кода.

  - Получает запрос от клиента и отправляет его в канал `ch` для воркеров. Если канал занят, возвращает ошибку `ResourceExhausted`.

  - Ожидает ответ от воркеров в виде JSON-подобной строки, затем десериализует и возвращает результат. Загружает результат в кэш.

5. метод-перехватчик **`LogUnaryRPCInterceptor`:**
```go
func (es *ExtServer) LogUnaryRPCInterceptor(ctx context.Context, req interface{},
	_ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error)
```
  - Перехватчик для логирования информации о каждом gRPC вызове.

  - Измеряет время выполнения вызова и логирует различную информацию о вызове, включая наличие данных в кэше.

  - Если данные найдены в кэше, возвращает их сразу, иначе передает выполнение обработчику.

6. метод **`StartServer`:**
```go
func (es *ExtServer) StartServer(addr, port string) (*grpc.Server, error)
```
  - Метод для запуска gRPC сервера.

  - Создает слушатель на указанном адресе и порту.

  - Создает новый экземпляр gRPC сервера с указанными опциями и регистрирует его методы.

  - Запускает сервер в горутине.

  - Возвращает указатель на сервер и ошибку (если есть).

Этот код представляет собой реализацию gRPC-сервиса, взаимодействующего с воркерами через канал и использующего логгирование с использованием библиотеки zap.
## _**Protobuf file**_ сервиса 
```protobuf
syntax = "proto3";

package service;

option go_package = "github.com/EgorKo25/DES/internal/server";

service UserExtensionService  {
  rpc GetUserExtension(GetRequest) returns (GetResponse);
}

message GetRequest {
  UserData user_data = 1;
}

message GetResponse {
  string status = 1;
  UserData users = 2;
}

message UserData {
  int32 ids = 1;
  string name = 2;
  string email = 3;
  string phone_number = 4;
  string date_to = 5;
  string date_from = 6;
}
```

## [Worker package](https://github.com/EgorKo25/DES/blob/main/internal/workers/worker.go)
1. Функция-конструктор **`NewWorkerPull`:**
    ```go
    NewWorkerPull(ctx context.Context, channel chan chan []byte, maxWorkers, timeOutConn, maxResponseTime int,
	login, password string, logger, htpLogger *zap.Logger) *WorkerPull
   ```
    - Инициализирует и возвращает новый экземпляр `WorkerPull`.
    - Запускает указанное количество воркеров в горутинах.
    - Устанавливает контекст, таймаут подключения, логин, пароль и другие параметры.
    - Проверяет, что контекст не содержит ошибку.

2. Структура **`WorkerPull`:**
```go
   type WorkerPull struct {
	*Auth

	channel    chan chan []byte
	logger     *zap.Logger
	httpLogger *zap.Logger
	client     *http.Client

	urlExtData   string
	urlStatusAbv string

	workerPullSize  int
	maxResponseTime int
}
``` 
   - Содержит информацию о воркере: 
     - аутентификацию **`(*Auth)`**
     - логгеры **`logger, httpLogger`**
     - клиент HTTP **`client`**
     - адреса обработчиков удаленного HTTP **`urlExtData, urlStatusAbv`**
     - Максимальное количесвто воркеров **`workerPullSize`**
     - Максимально время для ответа в канал **`maxResponseTime`**
   - Метод `worker` - выполняет обработку запросов из канала. Обрабатывает полученные данные, отправляет запросы и возвращает результаты.

3. **`Auth` структура:**
```go
type Auth struct {
	login    string
	password string
}
```

- Хранит логин и пароль для аутентификации.

4. **Методы `worker` и `processRequest`:**
+ `worker`: Обрабатывает данные из канала, отправляет запросы и возвращает результаты.
```go
func (wp *WorkerPull) worker(ctx context.Context)
```
+ `processRequest`: Отправляет HTTP-запрос с предоставленными данными и возвращает тело ответа.
```go
func (wp *WorkerPull) processRequest(ctx context.Context, url string, data *fastjson.Value, body []byte) ([]byte, error)
```
5. **`setSmile` функция:**
```go
func (wp *WorkerPull) setSmile(reasonId int) string 
```
+ Возвращает эмоджи в зависимости от переданного `reasonId`.

6. **`getAuthorization` функция:**
```go
func (wp *WorkerPull) getAuthorization() string
```

+ Возвращает строку для заголовка авторизации, созданную на основе логина и пароля.

7. **Логирование:**
    - Используется пакет `go.uber.org/zap` для логирования.
    - Различные логи, такие как информация о начале работы воркера, отправке запросов и получении ответов, сохранены для отслеживания действий приложения.

8. **Использование emoji:**
    - Используется пакет `github.com/enescakir/emoji` для добавления эмоджи в строку в зависимости от `reasonId`.

Этот код представляет собой воркер-пул для обработки запросов. Каждый воркер получает данные из канала, отправляет запросы на указанные URL-ы, обрабатывает ответы и возвращает результаты.
