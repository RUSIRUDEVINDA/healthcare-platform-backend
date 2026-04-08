package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// Verifies PUT /doctors/:id does not hit UpdatePutRoot (PUT /doctors).
func TestGinRoute_PUTDoctorsWithID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	var which string
	protected := r.Group("/doctors")
	protected.Use(func(c *gin.Context) { c.Next() }) // mirror extra middleware chain depth
	{
		protected.PUT("/:id/profile", func(c *gin.Context) { which = "profile" })
		protected.PUT("/:id", func(c *gin.Context) { which = "id:" + c.Param("id") })
		protected.PUT("", func(c *gin.Context) { which = "root" })
	}

	req := httptest.NewRequest(http.MethodPut, "/doctors/5", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if which != "id:5" {
		t.Fatalf("expected id:5 handler, got %q", which)
	}
}
