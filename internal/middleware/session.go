package middleware

import (
	"context"
	"net/http"
	"strings"
)

func SessionMetadata(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		ipAddress := getClientIP(r)
		ctx = context.WithValue(ctx, IPAddressKey, ipAddress)

		userAgent := r.Header.Get("User-Agent")
		ctx = context.WithValue(ctx, UserAgentKey, userAgent)

		deviceInfo := parseDeviceInfo(userAgent)
		ctx = context.WithValue(ctx, DeviceInfoKey, deviceInfo)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func parseDeviceInfo(userAgent string) string {
	if userAgent == "" {
		return "Unknown"
	}

	userAgent = strings.ToLower(userAgent)

	if strings.Contains(userAgent, "iphone") {
		return "iPhone"
	}
	if strings.Contains(userAgent, "ipad") {
		return "iPad"
	}
	if strings.Contains(userAgent, "android") {
		if strings.Contains(userAgent, "mobile") {
			return "Android Phone"
		}
		return "Android Tablet"
	}

	if strings.Contains(userAgent, "windows") {
		return "Windows PC"
	}
	if strings.Contains(userAgent, "macintosh") || strings.Contains(userAgent, "mac os x") {
		return "Mac"
	}
	if strings.Contains(userAgent, "linux") {
		return "Linux PC"
	}

	return "Desktop"
}
