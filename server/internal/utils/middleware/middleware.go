package middleware

import (
	"context"
	"os"
	"strings"

	"github.com/danielgtaylor/huma/v2"
)

func GetIpExtractor() func(ctx huma.Context, next func(huma.Context)) {
	// Get proxy IP address header from environment
	http_header := os.Getenv("MELO_WEBAPI_REAL_IP_HEADER")

	// Create closure for client IP extract
	return func(ctx huma.Context, next func(huma.Context)) {
		if ip := ctx.Header(http_header); ip != "" {
			ctx = huma.WithValue(ctx, "remote-ip", ip)
		} else {
			ctx = huma.WithValue(ctx, "remote-ip", strings.Split(ctx.RemoteAddr(), ":")[0])
		}
		next(ctx)
	}
}

func ExtractIp(ctx context.Context) string {
	// Get IP address of the remote
	ip, _ := ctx.Value("remote-ip").(string)
	return ip
}
