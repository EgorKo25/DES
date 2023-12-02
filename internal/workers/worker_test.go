package workers

import (
	"context"
	"fmt"
	"io"
	"log"
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

		client := NewWorkerPull(ctx, channel, maxWorkers, timeOutConn, 3, sugar)

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

		client := NewWorkerPull(ctx, channel, maxWorkers, timeOutConn, 3, sugar)
		require.Nil(t, client)
	})

	t.Log("Проверка функциональности при разных входных значениях при помощи эхо-сервера")
	type in struct {
		resonalId     int
		authorization string
		channel       chan chan []byte
		// UserDate
		email    string
		dateTo   string
		dateFrom string
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
			},
		},
	}
	//Конфигурация пула воркеров
	channel := make(chan chan []byte, 6)
	maxWorkers := 3
	timeOutConn := 0
	server := NewTestServer()
	defer server.Close()
	ctx, cancel := context.WithCancel(context.Background())

	for i, tt := range tests {
		t.Run(fmt.Sprintf("test №%d - %s", i, tt.name), func(t *testing.T) {

			resonalIds = tt.in.resonalId

			client := NewWorkerPull(ctx, channel, maxWorkers, timeOutConn, 3, sugar)
			require.NotNilf(t, client, "worker pull is nil!")

			client.urlExtData = server.URL + "/Portal/springApi/api/employees"
			client.urlStatusAbv = server.URL + "/Portal/springApi/api/absences"

			resultChan := make(chan []byte)
			channel <- resultChan

			resultChan <- []byte(fmt.Sprintf(`{"email":"%s", "dateTo":"%s", "dateFrom":"%s"}`, tt.in.email, tt.in.dateTo, tt.in.dateFrom))

			v := fastjson.MustParseBytes(<-resultChan)
			close(resultChan)
			cancel()

			log.Println(v.String())
			require.Equal(t, tt.want.status, v.Get("status").String())
			require.Equal(t, tt.want.id, v.GetInt("id"))
			require.Equal(t, tt.want.email, v.Get("email").String())
			require.Equal(t, tt.want.displayName, v.Get("displayName").String())
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

		//TODO:
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
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		answ := fmt.Sprintf(`{"status":"OK",
					"data":[
								{"id":1234,
								"displayName":"Иванов Семен Петрович",
								"email":"petrovich@mail.ru",
								"email":%s,
								"workPhone":"1234"}
							]}`, v.Get("email"))
		w.Write([]byte(answ))
	})

	server := httptest.NewServer(mux)
	return server
}
