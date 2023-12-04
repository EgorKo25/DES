package service

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"google.golang.org/grpc"

	pb "github.com/EgorKo25/DES/internal/proto/extension-service-gen"

	"github.com/stretchr/testify/require"
)

func TestExtServer_GetUserExtension(t *testing.T) {
	channel := make(chan chan []byte, 6)
	es := NewExtServer(channel)
	s := es.StartServer(":8080")
	{
		t.Run("Test Canceled Context", func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())

			conn, err := grpc.DialContext(ctx, ":8080", grpc.WithInsecure())
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
		}

		ctx := context.Background()

		conn, err := grpc.DialContext(ctx, ":8080", grpc.WithInsecure())
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
				log.Println(resp)
				require.NotNil(t, resp.Users)
				require.Equal(t, tt.want.ud.Name, resp.Users.Name)
			})
		}
		s.Stop()
	}
}
