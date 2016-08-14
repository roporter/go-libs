package headers

import (
	"github.com/kataras/iris"
)

var myVersion = ""

func Serve(ctx *iris.Context) {
	defer func() {
		ctx.Response.Header.Add("X-App-Version", myVersion)
	}()
	ctx.Next()
}

func New(version string) iris.HandlerFunc {
	myVersion = version
	return Serve
}
