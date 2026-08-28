package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// BodyLimit acota el cuerpo de la petición antes de que llegue al binding.
// Sin esto, ShouldBindJSON bufferizaba en memoria todo lo que mandara el
// cliente: un POST de cientos de megas contra un endpoint público bastaba
// para tumbar el proceso.
func BodyLimit(max int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, max)
		c.Next()
	}
}
