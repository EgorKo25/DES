# Date Extension Service (DES) 

**Сервис** занимаеться обогощением данных, получаемых по средствам <br> gRPC запросов.

## Как его собрать/запустить?  
```bash
    # чтобы запустить файл
    go run cmd/des/des.go
```
<br>

```bash 
    # чтобы собрать исполняемый файл
    go build cmd/des/des.go
    # запуск исполняемого файла
    ./des
```

## ![Typing SVG](https://readme-typing-svg.herokuapp.com?color=%2336BCF7&lines=Что+реализовано+в+проекте?)

# Пакет Workers

Пакет `workers` предоставляет простую реализацию пула горутин для параллельной обработки задач. Он разработан для выполнения задач с использованием горутин.

## Client

Структура `Client` представляет клиента для управления рабочими задачами. Он поддерживает создание рабочих горутин и обработку задач параллельно.

### Конструктор

```go
func NewClient(ctx context.Context, channel chan chan []byte, maxWorkers, timeOutConn, maxResponseTime int, logger Logger) *Client
```
Создает новый экземпляр Client с заданной конфигурацией.

* ctx: Контекст для управления жизненным циклом клиента.
* channel: Канал для получения задач.
* maxWorkers: Максимальное количество рабочих горутин.
* timeOutConn: Таймаут соединения для HTTP-запросов.
* maxResponseTime: Максимальное время ответа для HTTP-запросов.
* logger: Интерфейс логгера для ведения журнала.

### Worker
  Метод worker являеться типовым обработчиком, слушающим канал ```chan chan []byte```.

```go
func (w *Client) worker(ctx context.Context)
````
Он обрабатывает задачи из входного канала, выполняет HTTP-запросы и расширяет данные задачи соответственно.
### Обработка запроса
Метод processRequest используется для обработки HTTP-запросов.

```go
func (w *Client) processRequest(ctx context.Context, url string, data *fastjson.Value, body []byte) ([]byte, error)
```
Он создает HTTP-запрос, отправляет его и возвращает тело ответа.
### Установка смайла
Метод **setSmile** добавляет эмодзи в поле **displayName** на основе reasonId.

```go
func (w *Client) setSmile(reasonId int) string
```
Возвращает строковое представление эмодзи на основе указанного _**reasonId**_.
### Логгер
Интерфейс Logger определяет методы для ведения журнала сообщений различных уровней.
```go
type Logger interface {
    Infof(msg string, fields ...any)
    Errorf(msg string, fields ...any)
    Warnf(msg string, fields ...any)
    Debugf(msg string, fields ...any)
}
```
Он включает методы для ведения журнала информационных, ошибочных, предупредительных и отладочных сообщений.

### Ошибки
Пакет определяет константы для общих сообщений об ошибках:

+ **keepAliveConnectionTerminatError**: Сообщение об ошибке отсоединения keep-alive соединения.
+ **cannotReadBodyError**: Сообщение об ошибке чтения тела ответа.
+ **clientError**: Сообщение об ошибке, связанной с клиентом.
+ **requestCreateError**: Сообщение об ошибке создания HTTP-запроса.
+ **httpError**: Сообщение о получении HTTP ответа с ошибой.

