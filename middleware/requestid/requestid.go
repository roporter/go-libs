package requestid

import (
	"math/rand"
	"encoding/hex"
	"github.com/kataras/iris"
)

const Name = "RequestID"

func Serve(ctx *iris.Context) {
	defer func() {
		ctx.Response.Header.Add("X-Request-Id", uuid())
	}()
	ctx.Next()
}

func New() iris.HandlerFunc {
	return Serve
}

func uuid() string {
	u := make([]byte, 16)
	_, err := rand.Read(u)
	if err != nil {
		return ""
	}

	// this make sure that the 13th character is "4"
	u[6] = (u[6] | 0x40) & 0x4F
	// this make sure that the 17th is "8", "9", "a", or "b"
	u[8] = (u[8] | 0x80) & 0xBF 

	return hex.EncodeToString(u)
}
