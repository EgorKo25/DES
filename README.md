# Date Extension Service (DES) 

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

## [Config package](https://github.com/EgorKo25/DES/blob/config/internal/config/config.go)

### Описание

```go
package config

type AppConfig struct {
	WorkerConfig  WorkerConfig  `json:"worker"`
	ServiceConfig ServiceConfig `json:"service"`
}

type ServiceConfig struct {
	IP   string `json:"IP"`
	PORT string `json:"PORT"`
}

type WorkerConfig struct {
	MaxWorkers         int `json:"max_workers"`
	MaxTimeForResponse int `json:"max_time_for_response"`
	TimeoutConnection  int `json:"timeout_connection"`

	RemoteHTTPServer struct {
		IP   string `json:"IP"`
		PORT string `json:"PORT"`
	} `json:"remote_http_server"`
	Authentication struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	} `json:"authentication"`
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
}

```
## [Logger package](https://github.com/EgorKo25/DES/blob/config/internal/logger/logger.go)

Этот пакет предоставляет простой и гибкий механизм логгирования для вашего приложения на основе библиотеки [Zap](https://pkg.go.dev/go.uber.org/zap).

### Использование


### Конфигурация логгеров

- **Консольный вывод:**
  Логгер выводит сообщения на консоль с использованием консольного кодировщика для более читаемого вывода.

- **Файловый вывод:**
  Логгер также записывает сообщения в файлы для отслеживания деталей работы приложения. Различаются логи по уровням: `debug.log` для отладочных сообщений, `warning_error.log` для предупреждений и ошибок, а также `http.log` для отслеживания всех запросов к стороннему **HTTP** и `grpc.log` для отслеживания запросов по **grpc**.

- **Уровни логгирования:**
  Логгер настроен для разделения сообщений на уровни отладки, информации, предупреждений и ошибок.


Этот пакет предоставляет стандартный способ создания и настройки логгеров в вашем приложении, обеспечивая удобство использования и гибкость в конфигурации.