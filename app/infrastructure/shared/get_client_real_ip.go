package infrashared

import (
	"github.com/valyala/fasthttp"
)

// global variable client ip
var ClientIP *ClientRealIP

type ClientRealIP struct {
	IP string
}

func GetClientRealIP(ctx *fasthttp.RequestHeader) string {
	checklist := [...]string{
		"RPlus-Real-IP",
		"ali-cdn-real-ip",
		"CF-Connecting-IP",
		"X-Forwarded-For",
		"X-Forwarded",
		"X-Original-Forwarded-For",
		"Forwarded-For",
		"Forwarded",
		"True-Client-Ip",
		"X-Client-IP",
		"Fastly-Client-Ip",
		"X-Real-IP",
	}
	var result string
	for _, check := range checklist {
		if v := ctx.Peek(check); len(v) > 0 {
			result = string(v)
			break
		}
	}
	return result
}
