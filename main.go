package main

import (
	"net/http"
	"os"
	"github.com/mediocregopher/radix.v2/pool"
	"github.com/mediocregopher/radix.v2/redis"
)

// ./FBoxLiHTTPd ":8080" "127.0.0.1:6379" "redisdb (def: 0)" "keyprefix (def: fboxli:)" "redispw (def: none)"

type FBoxLiHandler struct {
	redisPool *pool.Pool
	redisPrefix string
}

func (h *FBoxLiHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain")

	res := h.redisPool.Cmd("GET", h.redisPrefix + r.URL.Path[1:])
	if res.Err != nil {
		w.WriteHeader(500)
		w.Write([]byte("500 - Internal Server Error"))
		return
	}

	url, err := res.Str()
	if err != nil || url == "" {
		w.WriteHeader(404)
		w.Write([]byte("404 - Not Found"))
		return
	}

	w.Header().Add("Location", url)
	w.WriteHeader(302)
	w.Write([]byte("Redirecting to: " + url))
}

func main() {
	argc := len(os.Args)

	redisDB := "0"
	redisPW := os.Getenv("REDIS_PASSWORD")
	redisPrefix := "fboxli:"

	if argc >= 5 {
		redisPrefix = os.Args[4]
	}
	if argc >= 4 {
		redisDB = os.Args[3]
	}

	var redisPool *pool.Pool
	var err error
	if redisDB != "0" || redisPW != "" {
		df := func(network, addr string) (*redis.Client, error) {
			client, err := redis.Dial(network, addr)
			if err != nil {
				return nil, err
			}
			if redisPW != "" {
				if err = client.Cmd("AUTH", redisPW).Err; err != nil {
					client.Close()
					return nil, err
				}
			}
			if redisDB != "0" {
				if err = client.Cmd("USE", redisDB).Err; err != nil {
					client.Close()
					return nil, err
				}
			}
			return client, nil
		}
		redisPool, err = pool.NewCustom("tcp", os.Args[2], 10, df)
	} else {
		redisPool, err = pool.New("tcp", os.Args[2], 10)
	}

	if err != nil {
		panic(err)
	}

	http.ListenAndServe(os.Args[1], &FBoxLiHandler{
		redisPool: redisPool,
		redisPrefix: redisPrefix,
	})
}