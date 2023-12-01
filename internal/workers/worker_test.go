package workers

import (
	"context"
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

func TestGetClient(t *testing.T) {
	t.Log("Positive Test: Successful client creation\n" +
		"Verify that the function returns the correct client and does not return an error on successful client creation.")
	{
		ctx := context.Background()
		channel := make(chan chan []byte)
		maxWorkers := 5
		timeOutConn := 10

		client := NewWorkerPull(ctx, channel, maxWorkers, timeOutConn, 3, sugar)

		if client == nil {
			t.Fatal("Expected a non-nil client, got nil")
		}
	}
	t.Log("Error Test: Invalid Context\n" +
		"Check that the function returns an error if an invalid (canceled) context is passed")
	{
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		channel := make(chan chan []byte)
		maxWorkers := 5
		timeOutConn := 10

		client := NewWorkerPull(ctx, channel, maxWorkers, timeOutConn, 3, sugar)
		require.Nil(t, client)
	}
	t.Log("Проверка прмой функциональности при помощи testserver")
	{
		//TODO:Сделай из этого блочный тест
		// TODO:Вынеси Сервер в глобальную переменную и сделай его ЭХОМ
		channel := make(chan chan []byte, 6)
		maxWorkers := 3
		timeOutConn := 0

		mux := http.NewServeMux()
		mux.HandleFunc("/Portal/springApi/api/absences", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			answ := `{
						"status": "OK",
						"data": [
						{
							"id":28246,
							"personId":1234,
							"createdDate":"2023-08-14",
							"dateFrom":"2023-08-12T00:00:00",
							"dateTo":"2023-08-12T23:59:59",
							"reasonId":9
						}
					]}`
			w.Write([]byte(answ))
		})
		mux.HandleFunc("/Portal/springApi/api/employees", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			answ := `{"status":"OK",
					"data":[
								{"id":1234,
								"displayName":"Иванов Семен Петрович",
								"email":"petrovich@mail.ru",
								"workPhone":"1234"}
							]}`
			w.Write([]byte(answ))
		})
		server := httptest.NewServer(mux)
		defer server.Close()

		ctx, cancel := context.WithCancel(context.Background())

		client := NewWorkerPull(ctx, channel, maxWorkers, timeOutConn, 3, sugar)

		sugar.Infof("client init!")

		require.NotNilf(t, client, "client is nil!")

		client.urlExtData = server.URL + "/Portal/springApi/api/employees"
		client.urlStatusAbv = server.URL + "/Portal/springApi/api/absences"

		resultChan := make(chan []byte)
		channel <- resultChan

		resultChan <- []byte(`{"email":"petrovich@mail.ru","dataTo":"10.20.30","dataFrom":"1/2/3"}`)

		v := fastjson.MustParseBytes(<-resultChan)
		close(resultChan)
		cancel()

		log.Println(v)
		require.Equal(t, 1234, v.GetInt("id"))
		require.Equal(t, `"petrovich@mail.ru"`, v.Get("email").String())
		require.Equal(t, `"Иванов Семен Петрович🎩"`, v.Get("displayName").String())
	}
}
