package workers

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/enescakir/emoji"

	"github.com/valyala/fastjson"
)

func NewWorkerPull(ctx context.Context, channel chan chan []byte, maxWorkers, timeOutConn, maxResponseTime int,
	login, password string,
	logger, htpLogger *zap.Logger) *WorkerPull {
	w := WorkerPull{
		Auth: &Auth{
			login:    login,
			password: password,
		},
		logger:     logger,
		httpLogger: htpLogger,
		channel:    channel,

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
		logger.Error("context error is not nil",
			zap.String("error", ctx.Err().Error()),
		)
		return nil
	}

	for i := 0; i < w.workerPullSize; i++ {
		w.logger.Info("starting worker",
			zap.Int("worker number", i+1),
		)
		go w.worker(ctx)
	}
	return &w
}

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

type Auth struct {
	login    string
	password string
}

func (wp *WorkerPull) worker(ctx context.Context) {
	a := fastjson.Arena{}
	body := make([]byte, 0)

	for {
		select {
		case <-ctx.Done():
			break
		case childCh, ok := <-wp.channel:
			if !ok {
				wp.logger.Error("client got a nil channel")
				close(childCh)
				break
			}
			d, ok := <-childCh
			if !ok {
				wp.logger.Error("response channel is close!")
				continue
			}

			userExtendedData, err := fastjson.ParseBytes(d)
			if err != nil {
				childCh <- []byte(`{"status":"BAD REQUEST"}`)
				close(childCh)
				continue
			}

			dateTo := userExtendedData.Get("dateTo")
			dateFrom := userExtendedData.Get("dateFrom")

			childCtx, cancel := context.WithTimeout(ctx, time.Duration(wp.maxResponseTime)*time.Second)

			if body, err = wp.processRequest(childCtx, wp.urlExtData, userExtendedData, body); err != nil {
				wp.logger.Error("process request error",
					zap.String("error", err.Error()),
				)
				childCh <- []byte(fmt.Sprintf(`{"status":"%s"}`, err))
				close(childCh)
				cancel()
				continue
			}

			userExtendedData, err = fastjson.ParseBytes(body)
			if err != nil {
				childCh <- []byte(fmt.Sprintf(`{"status":"%s"}`, err))
				close(childCh)
				continue
			}

			data := userExtendedData.GetArray("data")[0]
			data.Set("personalIds", a.NewArray())
			data.Get("personIds").SetArrayItem(0, data.Get("id"))
			data.Set("dateFrom", dateFrom)
			data.Set("dateTo", dateTo)

			if body, err = wp.processRequest(childCtx, wp.urlStatusAbv, data, body); err != nil {
				wp.logger.Error("process request error",
					zap.String("error", err.Error()),
				)
				childCh <- []byte(fmt.Sprintf(`{"status":"%s"}`, err))
				close(childCh)
				continue
			}
			userExtendedData, err = fastjson.ParseBytes(body)
			if err != nil {
				childCh <- []byte(`{"status":"INTERNAL SERVER ERROR"}`)
				close(childCh)
				continue
			}

			data.Del("personalIds")

			data.Set("displayName", a.NewString(string(data.GetStringBytes("displayName"))+
				wp.setSmile(userExtendedData.GetArray("data")[0].GetInt("reasonId"))))

			result := a.NewObject()
			result.Set("status", a.NewString("OK"))
			result.Set("users", data)

			childCh <- result.MarshalTo(d[:0])
			close(childCh)

			a.Reset()
			cancel()
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
		return nil, err
	}

	req.Header.Set("Authorization", wp.getAuthorization())

	timeStart := time.Now()

	resp, err := wp.client.Do(req)
	if err != nil {
		return nil, err
	}

	duration := time.Since(timeStart)

	if resp.StatusCode != http.StatusOK {
		switch resp.StatusCode {
		case http.StatusUnauthorized:
			return nil, errors.New("UNAUTHORIZED")
		case http.StatusBadRequest:
			return nil, errors.New("BAD REQUEST")
		}
		return nil, errors.New("INTERNAL SERVER ERROR")
	}

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if _, err = io.Copy(io.Discard, resp.Body); err != nil {
		return nil, err
	}

	wp.httpLogger.Info("http request done",
		zap.String("request", data.String()),
		zap.String("response code", resp.Status),
		zap.String("response", string(body)),
		zap.Duration("duration", duration),
	)

	return body, nil
}

func (wp *WorkerPull) getAuthorization() string {
	return base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", wp.login, wp.password)))
}
