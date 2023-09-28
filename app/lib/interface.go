package lib

import (
	"io"

	"github.com/gin-gonic/gin"
)

type Foreground interface {
	Start(gin.ResponseWriter) error
}

type Background interface {
	Start() error
	io.ReadCloser
}
