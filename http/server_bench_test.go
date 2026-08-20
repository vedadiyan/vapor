package http

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vedadiyan/vapor"

	"github.com/gin-gonic/gin"
	"github.com/gofiber/fiber/v2"
)

func BenchmarkHandleFunc(b *testing.B) {
	srv := New()

	if err := srv.HandleFunc(vapor.Pattern("GET /users/:id"), func(r vapor.Request) vapor.Response {
		return vapor.NewResponse(
			http.StatusOK,
			vapor.WithContent([]byte("hello")),
		)
	}); err != nil {
		b.Fatal(err)
	}

	s := srv.(*server)

	req := httptest.NewRequest(
		http.MethodGet,
		"/users/42",
		nil,
	)

	b.ReportAllocs()

	for b.Loop() {
		rec := httptest.NewRecorder()
		s.mux.ServeHTTP(rec, req)
	}
}

func BenchmarkRequestParams(b *testing.B) {
	srv := New()

	var req vapor.Request

	if err := srv.HandleFunc(vapor.Pattern("GET /users/:id"), func(r vapor.Request) vapor.Response {
		req = r
		return nil
	}); err != nil {
		b.Fatal(err)
	}

	s := srv.(*server)

	httpReq := httptest.NewRequest(
		http.MethodGet,
		"/users/42",
		nil,
	)

	s.mux.ServeHTTP(httptest.NewRecorder(), httpReq)

	b.ReportAllocs()

	for b.Loop() {
		_ = req.Params()
	}
}

func BenchmarkRequestHeaders(b *testing.B) {
	srv := New()

	var req vapor.Request

	if err := srv.HandleFunc(vapor.Pattern("GET /"), func(r vapor.Request) vapor.Response {
		req = r
		return nil
	}); err != nil {
		b.Fatal(err)
	}

	s := srv.(*server)

	httpReq := httptest.NewRequest(
		http.MethodGet,
		"/",
		nil,
	)

	httpReq.Header.Set("X-ID", "123")
	httpReq.Header.Set("X-Method", "GET")
	httpReq.Header.Set("Authorization", "Bearer token")
	httpReq.Header.Set("Content-Type", "application/json")

	s.mux.ServeHTTP(httptest.NewRecorder(), httpReq)

	b.ReportAllocs()

	for b.Loop() {
		_ = req.Headers()
	}
}

func BenchmarkRequestContent(b *testing.B) {
	srv := New()

	var req vapor.Request

	if err := srv.HandleFunc(vapor.Pattern("POST /"), func(r vapor.Request) vapor.Response {
		req = r
		return nil
	}); err != nil {
		b.Fatal(err)
	}

	s := srv.(*server)

	httpReq := httptest.NewRequest(
		http.MethodPost,
		"/",
		nil,
	)

	s.mux.ServeHTTP(httptest.NewRecorder(), httpReq)

	b.ReportAllocs()

	for b.Loop() {
		body := req.Content()
		_, _ = io.ReadAll(body)
		body.Close()
	}
}

func BenchmarkRequestMethod(b *testing.B) {
	srv := New()

	var req vapor.Request

	if err := srv.HandleFunc(vapor.Pattern("GET /"), func(r vapor.Request) vapor.Response {
		req = r
		return nil
	}); err != nil {
		b.Fatal(err)
	}

	s := srv.(*server)

	httpReq := httptest.NewRequest(
		http.MethodGet,
		"/",
		nil,
	)

	s.mux.ServeHTTP(httptest.NewRecorder(), httpReq)

	b.ReportAllocs()

	for b.Loop() {
		_ = req.Method()
	}
}

func BenchmarkFullRequest(b *testing.B) {
	srv := New()

	if err := srv.HandleFunc(vapor.Pattern("GET /users/:id"), func(r vapor.Request) vapor.Response {
		_ = r.Method()
		_ = r.ID()
		_ = r.Params()
		_ = r.QueryString()
		_ = r.Headers()

		return vapor.NewResponse(
			http.StatusOK,
			vapor.WithContent([]byte("hello")),
		)
	}); err != nil {
		b.Fatal(err)
	}

	s := srv.(*server)

	b.ReportAllocs()

	for b.Loop() {
		req := httptest.NewRequest(
			http.MethodGet,
			"/users/42?foo=bar",
			nil,
		)
		req.Header.Set("X-ID", "abc")
		req.Header.Set("X-Method", "GET")

		rec := httptest.NewRecorder()
		s.mux.ServeHTTP(rec, req)
	}
}

func BenchmarkNativeFullRequest(b *testing.B) {
	var mux http.ServeMux

	mux.HandleFunc("GET /users/{id}", func(w http.ResponseWriter, r *http.Request) {
		_ = r.Method
		_ = r.Header.Get("X-ID")
		_ = r.PathValue("id")
		_ = r.URL.RawQuery
		_ = r.Header

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("hello"))
	})

	b.ReportAllocs()

	for b.Loop() {
		req := httptest.NewRequest(
			http.MethodGet,
			"/users/42?foo=bar",
			nil,
		)
		req.Header.Set("X-ID", "abc")
		req.Header.Set("X-Method", "GET")

		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
	}
}

func BenchmarkGinFullRequest(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()

	r.GET("/users/:id", func(c *gin.Context) {
		_ = c.Request.Method
		_ = c.GetHeader("X-ID")
		_ = c.Param("id")
		_ = c.Request.URL.RawQuery
		_ = c.Request.Header

		c.String(http.StatusOK, "hello")
	})

	b.ReportAllocs()

	for b.Loop() {
		req := httptest.NewRequest(
			http.MethodGet,
			"/users/42?foo=bar",
			nil,
		)
		req.Header.Set("X-ID", "abc")
		req.Header.Set("X-Method", "GET")

		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
	}
}
func BenchmarkFiberFullRequest(b *testing.B) {
	app := fiber.New()

	app.Get("/users/:id", func(c *fiber.Ctx) error {
		_ = c.Method()
		_ = c.Get("X-ID")
		_ = c.Params("id")
		_ = c.Context().QueryArgs().String()

		return c.Status(http.StatusOK).SendString("hello")
	})

	req := httptest.NewRequest(
		http.MethodGet,
		"/users/42?foo=bar",
		nil,
	)
	req.Header.Set("X-ID", "abc")
	req.Header.Set("X-Method", "GET")

	b.ReportAllocs()

	for b.Loop() {
		rec := httptest.NewRecorder()

		resp, err := app.Test(req, -1)
		if err != nil {
			b.Fatal(err)
		}

		_, _ = io.Copy(rec, resp.Body)
		_ = resp.Body.Close()
	}
}
