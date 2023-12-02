package workers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/enescakir/emoji"

	"github.com/valyala/fastjson"
)

//TODO: README FIX

type Logger interface {
	Infof(msg string, fields ...any)
	Errorf(msg string, fields ...any)
	Warnf(msg string, fields ...any)
	Debugf(msg string, fields ...any)
}

var (
	keepAliveConnectionTerminatError = "error keep alive"
	cannotReadBodyError              = "cannot read response body"
	clientError                      = "client cannot send request"
	requestCreateError               = "cannot create new request"
	httpError                        = "not allowed response code from http"
)

type WorkerPull struct {
	channel chan chan []byte
	logger  Logger
	client  *http.Client

	urlExtData   string
	urlStatusAbv string

	workerPullSize  int
	maxResponseTime int
}

func NewWorkerPull(ctx context.Context, channel chan chan []byte, maxWorkers, timeOutConn, maxResponseTime int, logger Logger) *WorkerPull {
	w := WorkerPull{
		logger:          logger,
		channel:         channel,
		maxResponseTime: maxResponseTime,
		workerPullSize:  maxWorkers,
		client: &http.Client{
			Timeout: time.Duration(timeOutConn) * time.Second,
			Transport: &http.Transport{
				IdleConnTimeout: 0,
			},
		},
	}

	if ctx.Err() != nil {
		return nil
	}

	for i := 0; i < w.workerPullSize; i++ {
		w.logger.Infof("worker %d is running", i+1)
		go w.worker(ctx)
	}

	return &w
}

func (wp *WorkerPull) worker(ctx context.Context) {
	a := fastjson.Arena{}
	body := make([]byte, 0)

	var err error

	for {
		select {
		case <-ctx.Done():
			break
		case childCh, ok := <-wp.channel:

			if !ok {
				wp.logger.Infof("client got a nil channel")
				break
			}
			d, ok := <-childCh
			if !ok {
				wp.logger.Errorf("response channel is close!")
				childCh <- []byte(`{"status":"CLOSE CHAN"}`)
				continue
			}

			userExtendedData := fastjson.MustParseBytes(d)

			dateTo := userExtendedData.Get("dateTo")
			dateFrom := userExtendedData.Get("dateFrom")

			childCtx, cancel := context.WithTimeout(ctx, time.Duration(wp.maxResponseTime)*time.Second)

			if body, err = wp.processRequest(childCtx, wp.urlExtData, userExtendedData, body); err != nil {
				childCh <- []byte(fmt.Sprintf(`{"status":"%s"}`, err))
				cancel()
				continue
			}

			userExtendedData = fastjson.MustParseBytes(body)

			data := userExtendedData.GetArray("data")[0]
			data.Set("personalIds", a.NewArray())
			data.Get("personIds").SetArrayItem(0, data.Get("id"))
			data.Set("dateFrom", dateFrom)
			data.Set("dateTo", dateTo)

			if body, err = wp.processRequest(childCtx, wp.urlStatusAbv, data, body); err != nil {
				wp.logger.Errorf("%v", err)
				childCh <- []byte(`{"status":"INTERNAL SERVER ERROR"}`)
				continue
			}
			//TODO:Обработка ответов стороннего сервера
			//TODO:НЕизвестный ответ сервера
			userExtendedData, err = fastjson.ParseBytes(body)
			if err != nil {
				childCh <- []byte(`{"status":"INTERNAL SERVER ERROR"}`)
				continue
			}

			data.Del("personalIds")

			data.Set("displayName", a.NewString(string(data.GetStringBytes("displayName"))+
				wp.setSmile(userExtendedData.GetArray("data")[0].GetInt("reasonId"))))
			data.Set("status", a.NewString("OK"))

			childCh <- data.MarshalTo(d[:0])

			a.Reset()
		default:
		}
	}
}

func (wp *WorkerPull) setSmile(reasonId int) string {
	switch reasonId {
	case 1, 10:
		return string(emoji.House)
	case 3, 4:
		return string(emoji.Airplane)
	case 5, 6:
		return string(emoji.Thermometer)
	case 9:
		return string(emoji.TopHat)
	case 11, 12, 13:
		return string(emoji.Sun)
	}
	return ""
}

func (wp *WorkerPull) processRequest(ctx context.Context, url string, data *fastjson.Value, body []byte) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		url, bytes.NewReader(data.MarshalTo(body[:0])))
	if err != nil {
		wp.logger.Errorf(requestCreateError)
		return nil, err
	}

	resp, err := wp.client.Do(req)
	if err != nil {
		wp.logger.Errorf(clientError)
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		wp.logger.Errorf(httpError)
		return nil, errors.New(httpError)
	}

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		wp.logger.Errorf(cannotReadBodyError)
		return nil, err
	}

	if _, err = io.Copy(io.Discard, resp.Body); err != nil {
		wp.logger.Errorf(keepAliveConnectionTerminatError)
		return nil, err
	}

	return body, nil
}
