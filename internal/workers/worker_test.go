package workers

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/valyala/fastjson"

	"github.com/stretchr/testify/require"

	"go.uber.org/zap"
)

var logger, _ = zap.NewProduction()
var sugar = logger.Sugar()

// Эхо Переменные тест сервера
var (
	authorization string
	displayName   string
	resonalIds    int
)

func TestWorkerPull(t *testing.T) {
	t.Run("Положительный тест: Попытка просто создать клиент", func(t *testing.T) {
		t.Log("Клиенет создаёться с допустимыми параметрами")

		ctx := context.Background()
		channel := make(chan chan []byte)
		maxWorkers := 5
		timeOutConn := 10

		client := NewWorkerPull(ctx, channel, maxWorkers, timeOutConn, 3, sugar, "", "")

		if client == nil {
			t.Fatal("Expected a non-nil client, got nil")
		}
	})
	t.Run("Негативный тест: Нерабочий контекст", func(t *testing.T) {
		t.Log("Проверяет что вернет функция Если контекст будет отменён")

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		channel := make(chan chan []byte)
		maxWorkers := 5
		timeOutConn := 10

		client := NewWorkerPull(ctx, channel, maxWorkers, timeOutConn, 3, sugar, "", "")
		require.Nil(t, client)
	})

	t.Log("Проверка функциональности при разных входных значениях при помощи эхо-сервера")
	type in struct {
		resonalId     int
		authorization Auth
		channel       chan chan []byte
		// UserDate
		email    string
		dateTo   string
		dateFrom string
		// Config Test server
		displayName string
	}

	type want struct {
		status      string
		email       string
		id          int
		displayName string
	}
	tests := []struct {
		name string
		want want
		in   in
	}{
		{
			name: "Simple positive test case",
			want: want{
				status:      `"OK"`,
				email:       `"petrovich@mail.ru"`,
				id:          1234,
				displayName: `"Иванов Семен Петрович🎩"`,
			},
			in: in{
				resonalId: 9,

				email:    "petrovich@mail.ru",
				dateFrom: "10.20.30",
				dateTo:   "1/2/3",

				displayName: "Иванов Семен Петрович",
			},
		},
		{
			name: "Unauthorized test case",
			want: want{
				status: `"UNAUTHORIZED"`,
			},
			in: in{
				resonalId: 0,
				authorization: Auth{
					login:    "admin",
					password: "admin",
				},

				email:    "petrovich@mail.ru",
				dateFrom: "10.20.30",
				dateTo:   "1/2/3",

				displayName: "Иванов Семен Петрович",
			},
		},
		{
			name: "Request without email test case",
			want: want{
				status: `"BAD REQUEST"`,
			},
			in: in{
				resonalId: 0,
				authorization: Auth{
					login:    "admin",
					password: "admin",
				},

				displayName: "Иванов Семен Петрович",
			},
		},
		{
			name: "Wrong http response test case",
			want: want{
				status: `"INTERNAL SERVER ERROR"`,
			},
			in: in{
				resonalId: 0,
				email:     `"petrovich@mail.ru"`,

				displayName: `"Иванов"" ,}Семен "Петрович"`,
			},
		},
	}
	//Конфигурация пула воркеров
	channel := make(chan chan []byte, 6)
	maxWorkers := 3
	timeOutConn := 0
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Запуск тест вервера
	server := NewTestServer()
	defer server.Close()

	for i, tt := range tests {
		t.Run(fmt.Sprintf("test №%d - %s", i, tt.name), func(t *testing.T) {

			resonalIds = tt.in.resonalId
			displayName = tt.in.displayName
			authorization = base64.StdEncoding.EncodeToString([]byte(
				fmt.Sprintf("%s:%s", tt.in.authorization.login, tt.in.authorization.password)),
			)

			workerPull := NewWorkerPull(ctx, channel,
				maxWorkers, timeOutConn,
				3, sugar,
				tt.in.authorization.login,
				tt.in.authorization.password,
			)

			require.NotNilf(t, workerPull, "worker pull is nil!")

			workerPull.urlExtData = server.URL + "/Portal/springApi/api/employees"
			workerPull.urlStatusAbv = server.URL + "/Portal/springApi/api/absences"

			resultChan := make(chan []byte)
			channel <- resultChan

			resultChan <- []byte(fmt.Sprintf(`{"email":"%s", "dateTo":"%s", "dateFrom":"%s"}`, tt.in.email, tt.in.dateTo, tt.in.dateFrom))

			v := fastjson.MustParseBytes(<-resultChan)
			close(resultChan)

			require.Equal(t, tt.want.status, v.Get("status").String())
			if v.Exists("id", "email", "displayName") {
				require.Equal(t, tt.want.id, v.GetInt("id"))
				require.Equal(t, tt.want.email, v.Get("email").String())
				require.Equal(t, tt.want.displayName, v.Get("displayName").String())
			}
		})
	}
}

// NewTestServer возвращает эхо-серер для тестов
func NewTestServer() *httptest.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/Portal/springApi/api/absences", func(w http.ResponseWriter, r *http.Request) {

		auth := r.Header.Get("Authorization")

		if auth != authorization {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		v, err := fastjson.ParseBytes(body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		answ := fmt.Sprintf(`{
						"status": "OK",
						"data": [
						{
							"id":1234,
							"personId":%s,
							"createdDate":"20.20.20",
							"dateFrom":%s,
							"dateTo":%s,
							"reasonId":%d
						}
					]}`, v.Get("personalIds").String(), v.Get("dateFrom").String(), v.Get("dateTo").String(), resonalIds)
		w.Write([]byte(answ))
	})
	mux.HandleFunc("/Portal/springApi/api/employees", func(w http.ResponseWriter, r *http.Request) {

		auth := r.Header.Get("Authorization")
		if auth != authorization {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		v, err := fastjson.ParseBytes(body)
		if err != nil {
			sugar.Errorf("cannot read resp body - %s", v.String())
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if str := v.Get("email"); str.String() != "" {
			sugar.Errorf("have not key email in resp body - %s", v.String())
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		answ := fmt.Sprintf(`{"status":"OK",
					"data":[
								{"id":1234,
								"displayName":"%s",
								"email":%s,
								"workPhone":"1234"}
							]}`, displayName, v.Get("email"))
		w.Write([]byte(answ))
	})

	server := httptest.NewServer(mux)
	return server
}
