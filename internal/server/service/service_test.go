package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/EgorKo25/DES/internal/cache"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	pb "github.com/EgorKo25/DES/internal/server/extension-service-gen"

	"github.com/stretchr/testify/require"
)

func TestExtServer_GetUserExtension(t *testing.T) {

	logger := zap.NewExample()
	creds := insecure.NewCredentials()
	channel := make(chan chan []byte, 10)
	caches := cache.NewCache(context.Background(), 3)
	es := NewExtServer(channel, logger, logger, caches)
	s, _ := es.StartServer("", ":8080")
	{

		t.Run("Test Canceled Context", func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())

			conn, err := grpc.DialContext(ctx, ":8080", grpc.WithTransportCredentials(creds))
			if err != nil {
				require.Nil(t, err, fmt.Sprintf("error must to be nil: %s", err))
			}
			defer conn.Close()

			c := pb.NewUserExtensionServiceClient(conn)
			cancel()

			_, err = c.GetUserExtension(ctx, nil)

			require.Equal(t, "rpc error: code = Canceled desc = context canceled", err.Error())
		})
	}
	{
		type enter struct {
			ud                   *pb.UserData
			dataFromBlackBoxHTTP string
		}
		type want struct {
			code string
			ud   *pb.UserData
			err  string
		}
		tests := []struct {
			name  string
			enter enter
			want  want
		}{
			{
				name: "Test Internal Server Error",
				want: want{
					err: "rpc error: code = Internal desc = INTERNAL SERVER ERROR",
				},
			},
			{
				name: "Test Success Server Response",
				enter: enter{
					ud: &pb.UserData{
						Email:    "egorko@at.com",
						DateTo:   "10/20/30",
						DateFrom: "5/7/9",
					},
					dataFromBlackBoxHTTP: `{"status":"OK", "users":{"email":"egorko@at.com","name":"Егор Иванович"}}`,
				},
				want: want{
					code: "OK",
					err:  "",
					ud: &pb.UserData{
						Email:    "egorko@at.com",
						DateTo:   "10/20/30",
						DateFrom: "5/7/9",
						Name:     "Егор Иванович",
					},
				},
			},
			{
				name: "Test Success Server Response With Cache",
				enter: enter{
					ud: &pb.UserData{
						Email:    "egorko@at.com",
						DateTo:   "10/20/30",
						DateFrom: "5/7/9",
					},
				},
				want: want{
					code: "OK",
					err:  "",
					ud: &pb.UserData{
						Email:    "egorko@at.com",
						DateTo:   "10/20/30",
						DateFrom: "5/7/9",
						Name:     "Егор Иванович",
					},
				},
			},
		}

		ctx := context.Background()

		conn, err := grpc.DialContext(ctx, ":8080", grpc.WithTransportCredentials(creds))
		if err != nil {
			require.Nil(t, err, fmt.Sprintf("error must to be nil: %s", err))
		}
		defer conn.Close()

		c := pb.NewUserExtensionServiceClient(conn)

		for i, tt := range tests {
			t.Run(fmt.Sprintf("Test №%d: %s", i, tt.name), func(t *testing.T) {
				req := &pb.GetRequest{UserData: tt.enter.ud}

				childCtx, cancel := context.WithTimeout(ctx, 1*time.Second)
				defer cancel()

				//Симуляция ответа стороннего HTTP
				go func() {
					result := <-channel
					<-result
					result <- []byte(tt.enter.dataFromBlackBoxHTTP)
				}()

				resp, err := c.GetUserExtension(childCtx, req)
				if err != nil {
					require.Equal(t, tt.want.err, err.Error())
					return
				}

				require.NotNil(t, resp.Status)
				require.Equal(t, tt.want.code, resp.Status)
				require.NotNil(t, resp.Users)
				require.Equal(t, tt.want.ud.Name, resp.Users.Name)

			})
		}
	}
	{

		channel := make(chan chan []byte, 1)
		es := NewExtServer(channel, logger, logger, caches)
		_, _ = es.StartServer("", ":8080")

		t.Run("Test Too Many Request", func(t *testing.T) {
			ctx := context.Background()

			childCtx, cancel := context.WithTimeout(ctx, 1*time.Second)
			defer cancel()

			conn, err := grpc.DialContext(ctx, ":8080", grpc.WithTransportCredentials(creds))
			if err != nil {
				require.Nil(t, err, fmt.Sprintf("error must to be nil: %s", err))
			}
			defer conn.Close()

			c := pb.NewUserExtensionServiceClient(conn)

			go func() {
				_, _ = c.GetUserExtension(childCtx, &pb.GetRequest{UserData: &pb.UserData{Email: ""}})
			}()

			_, err = c.GetUserExtension(childCtx, &pb.GetRequest{UserData: &pb.UserData{Email: ""}})

			require.Equalf(t, status.Error(codes.ResourceExhausted, "TOO MANY REQUEST"), err, err.Error())
		})
	}
	s.Stop()
}
