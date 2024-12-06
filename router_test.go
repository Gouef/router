package router

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRouterErrorHandlers(t *testing.T) {
	// Nastavíme Gin do testovacího režimu
	gin.SetMode(gin.TestMode)

	// Vytvoření nového routeru
	router := NewRouter()

	// Nastavení vlastních handlerů pro testování
	router.SetErrorHandler(http.StatusNotFound, func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Custom 404",
		})
	})

	router.SetErrorHandler(http.StatusInternalServerError, func(c *gin.Context) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Custom 500",
		})
	})

	// Definice rout pro testování
	router.AddRouteGet("/ok", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "This is OK"})
	})
	router.AddRouteGet("/notfound", func(c *gin.Context) {
		c.Status(http.StatusNotFound)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Custom 404",
		})
	})
	router.AddRouteGet("/servererror", func(c *gin.Context) {
		c.Status(http.StatusInternalServerError)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Custom 500",
		})
	})

	// Test: `/ok` (status 200)
	t.Run("Test OK Status", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/ok", nil)
		w := httptest.NewRecorder()

		router.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, `{"message":"This is OK"}`, w.Body.String())
	})

	// Test: `/notfound` (status 404)
	t.Run("Test Custom 404 Handler", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/notfound", nil)
		w := httptest.NewRecorder()

		router.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.JSONEq(t, `{"error":"Custom 404"}`, w.Body.String())
	})

	// Test: `/servererror` (status 500)
	t.Run("Test Custom 500 Handler", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/servererror", nil)
		w := httptest.NewRecorder()

		router.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.JSONEq(t, `{"error":"Custom 500"}`, w.Body.String())
	})

	// Test: Neznámá cesta (status 404)
	t.Run("Test Unknown Route", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/unknown", nil)
		w := httptest.NewRecorder()

		router.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.JSONEq(t, `{"error":"Custom 404"}`, w.Body.String())
	})
}

func TestRouter_AddRoute(t *testing.T) {
	router := NewRouter()

	// DTO a handler
	type ProductDetailParams struct {
		Locale string `uri:"locale" binding:"required"`
		ID     int    `uri:"id" binding:"required"`
	}
	productDetailHandler := func(c *gin.Context, params *ProductDetailParams) {
		c.JSON(http.StatusOK, gin.H{
			"locale": params.Locale,
			"id":     params.ID,
		})

		assert.Equal(t, "cs", params.Locale)
		assert.Equal(t, 42, params.ID)
	}

	// Přidání routy
	router.AddRoute("/:locale/products/:id", productDetailHandler, Get)
	router.AddRoute(
		"/product/:id",
		func(c *gin.Context, p *struct {
			ID int `uri:"id" binding:"required"`
		}) {
			c.JSON(http.StatusOK, gin.H{
				"id": p.ID,
			})

			assert.Equal(t, 42, p.ID)
		}, Get)

	// Testování požadavku
	req := httptest.NewRequest(http.MethodGet, "/cs/products/42", nil)
	w := httptest.NewRecorder()
	router.GetNativeRouter().ServeHTTP(w, req)

	// Ověření výsledku
	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"locale":"cs","id":42}`, w.Body.String())

	// Testování požadavku
	req2 := httptest.NewRequest(http.MethodGet, "/product/42", nil)
	w2 := httptest.NewRecorder()
	router.GetNativeRouter().ServeHTTP(w2, req2)

	// Ověření výsledku
	assert.Equal(t, http.StatusOK, w2.Code)
	assert.JSONEq(t, `{"id":42}`, w2.Body.String())
}

func TestRouter_AddRouteWithoutParams(t *testing.T) {
	router := NewRouter()

	productDetailHandler := func(c *gin.Context) {
		l := c.Param("locale")
		id := c.Param("id")
		c.JSON(http.StatusOK, gin.H{
			"locale": l,
			"id":     id,
		})

		assert.Equal(t, "cs", l)
		assert.Equal(t, "42", id)
	}

	// Přidání routy
	router.AddRoute("/:locale/products/:id", productDetailHandler, Get)

	// Testování požadavku
	req := httptest.NewRequest(http.MethodGet, "/cs/products/42", nil)
	w := httptest.NewRecorder()
	router.GetNativeRouter().ServeHTTP(w, req)

	// Ověření výsledku
	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"locale":"cs","id":"42"}`, w.Body.String())
}

func TestNewRouteList(t *testing.T) {
	type ProductDetailParams struct {
		Locale string `uri:"locale" binding:"required"`
		ID     int    `uri:"id" binding:"required"`
	}
	productDetailHandler := func(c *gin.Context, params *ProductDetailParams) {
		c.JSON(http.StatusOK, gin.H{
			"locale": params.Locale,
			"id":     params.ID,
		})

		assert.Equal(t, "cs", params.Locale)
		assert.Equal(t, 42, params.ID)
	}

	lr := NewRouteList()
	v1 := CreateRouteList("/v1")
	lr.addChild(v1)

	lr.Add("/:locale/products/:id", productDetailHandler, Get)
	v1.Add("/:locale/products/:id", productDetailHandler, Get)

	router := NewRouter()
	router.AddRouteList(lr)
	CreateRoute(router, "/test/:id", func(c *gin.Context, p *struct {
		ID int `uri:"id" binding:"required"`
	}) {
		c.JSON(http.StatusOK, gin.H{
			"id": p.ID,
		})
		assert.Equal(t, 42, p.ID)
	}, Get)

	// Testování požadavku
	req := httptest.NewRequest(http.MethodGet, "/cs/products/42", nil)
	w := httptest.NewRecorder()
	router.GetNativeRouter().ServeHTTP(w, req)

	// Ověření výsledku
	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"locale":"cs","id":42}`, w.Body.String())

	// Testování požadavku
	req2 := httptest.NewRequest(http.MethodGet, "/v1/cs/products/42", nil)
	w2 := httptest.NewRecorder()
	router.GetNativeRouter().ServeHTTP(w2, req2)

	// Ověření výsledku
	assert.Equal(t, http.StatusOK, w2.Code)
	assert.JSONEq(t, `{"locale":"cs","id":42}`, w2.Body.String())

}
