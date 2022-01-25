package knocker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/turanukimaru/knocker/auth"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"testing"
)

func TestKnock(t *testing.T) {

	// デフォルトハンドラを使うときはこの形式
	http.HandleFunc("/ping", pongHandler) // ハンドラを登録してウェブページを表示させる

	server := http.Server{Addr: ":8080"}
	wg := sync.WaitGroup{}
	wg.Add(1)
	go ServerStart(&wg, &server)()
	k := Knocker{"127.0.0.1", 8080, ""}
	{
		v := &PongResponse{}
		res, b, err := k.Knock(http.MethodGet, "/ping", nil, v)
		if err != nil {
			panic(err)
		}
		assert.Equal(t, 200, res.StatusCode)
		fmt.Println(b)
		assert.Equal(t, "pong!", v.Res)
	}

	log.Printf("test done\n")

	if err := server.Shutdown(context.TODO()); err != nil {
		log.Printf("Shutdown error\n")
		log.Println(err)
		panic(err)
	}

	wg.Wait()
	log.Printf("receive server shutdown\n")
}

func TestPost(t *testing.T) {

	// デフォルトハンドラを使うときはこの形式
	http.HandleFunc("/ping", postHandler) // ハンドラを登録してウェブページを表示させる

	server := http.Server{Addr: ":8080"}
	wg := sync.WaitGroup{}
	wg.Add(1)
	go ServerStart(&wg, &server)()
	k := Knocker{"127.0.0.1", 8080, ""}
	{
		p := PongRequest{"↑", Content{"↓"}}
		j, err := json.Marshal(p)
		buffer := bytes.Buffer{}
		buffer.Write(j)
		v := &PongRequest{}
		res, b, err := k.Knock(http.MethodGet, "/ping", &buffer, v)
		if err != nil {
			panic(err)
		}
		assert.Equal(t, 200, res.StatusCode)
		fmt.Printf("response:%s\n", b)
		assert.Equal(t, "req=↑", v.Req)
	}

	log.Printf("test done\n")

	if err := server.Shutdown(context.TODO()); err != nil {
		log.Printf("Shutdown error\n")
		log.Println(err)
		panic(err)
	}

	wg.Wait()
	log.Printf("receive server shutdown\n")
}

func TestAuth(t *testing.T) {
	// 特に意味はないがデフォルトではないルータにしてみる
	r := mux.NewRouter()
	r.Handle("/auth", auth.GetTokenHandler)
	r.Handle("/private", auth.JwtMiddleware.Handler(private))

	server := http.Server{Addr: ":8080", Handler: r}
	wg := sync.WaitGroup{}
	wg.Add(1)
	go ServerStart(&wg, &server)()
	k := Knocker{"127.0.0.1", 8080, ""}
	t.Run("Required authorization token not found", func(t *testing.T) {
		res, b, err := k.Knock(http.MethodGet, "/private", nil, nil)
		if err != nil {
			panic(err)
		}
		assert.Equal(t, 401, res.StatusCode)
		assert.Equal(t, b, "Required authorization token not found\n")
	})
	t.Run("Auth and OK", func(t *testing.T) {
		res, _, err := k.Auth(http.MethodGet, "/auth", nil)
		if err != nil {
			panic(err)
		}
		v := &PrivateResponse{}
		res, _, err = k.Knock(http.MethodGet, "/private", nil, v)
		if err != nil {
			panic(err)
		}
		assert.Equal(t, 200, res.StatusCode)
		assert.Equal(t, v.Tag, "Go")
	})

	log.Printf("test done\n")

	if err := server.Shutdown(context.TODO()); err != nil {
		log.Printf("Shutdown error\n")
		log.Println(err)
		panic(err)
	}

	wg.Wait()
	log.Printf("receive server shutdown\n")
}

// Start server in go routine
func ServerStart(wg *sync.WaitGroup, server *http.Server) func() {
	return func() {
		defer func() {
			log.Printf("server shutdown\n")
			wg.Done()
		}()
		log.Printf("server start\n")
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Printf("ListenAndServe error\n")
			log.Println(err)
			panic(err)
		}
		log.Printf("server.Shutdown -> http: Server closed\n")
	}
}

// Public api
func postHandler(w http.ResponseWriter, r *http.Request) {
	resBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	v := PongRequest{}
	if err := json.Unmarshal(resBody, &v); err != nil {
		panic(err)
	}
	v.Req = "req=" + v.Req
	v.Child.Contents = "contents=" + v.Child.Contents
	res, err := json.Marshal(v)
	resString := string(res)
	if _, err := fmt.Fprintf(w, "%s", resString); err != nil {
		panic(err)
	}
}

// Public api
func pongHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := fmt.Fprintf(w, `{"res":"pong!"}`); err != nil {
		panic(err)
	}
}

// Pong Response
type PongResponse struct {
	Res string
}

// Pong Response
type PongRequest struct {
	Req   string
	Child Content
}
type Content struct {
	Contents string
}

// Private api
var private = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	reply := PrivateResponse{
		Title: "jwt token が無いとアクセスできないはず",
		Tag:   "Go",
	}
	json.NewEncoder(w).Encode(reply)
})

// Private Response
type PrivateResponse struct {
	Title string
	Tag   string
}
