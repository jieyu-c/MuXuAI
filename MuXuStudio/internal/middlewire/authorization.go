package middlewire

import (
	"context"

	"github.com/go-kratos/kratos/v3/middleware"
	khttp "github.com/go-kratos/kratos/v3/transport/http"
	"jieyuc.cn/muxu/studio/internal/identity"
)

// Route identifies a public endpoint by exact HTTP method and path.
// Query parameters are ignored; methods use uppercase, e.g. POST.
type Route struct {
	Method string
	Path   string
}

// Authorization exempts explicitly public HTTP routes and delegates all other
// requests to identity.Authenticate. Non-HTTP requests never match public routes.
func Authorization(verifier identity.TokenVerifier, publicRoutes ...Route) middleware.Middleware {
	public := make(map[Route]struct{}, len(publicRoutes))
	for _, route := range publicRoutes {
		public[route] = struct{}{}
	}
	authenticate := identity.Authenticate(verifier)
	return func(next middleware.Handler) middleware.Handler {
		protected := authenticate(next)
		return func(ctx context.Context, req any) (any, error) {
			if request, ok := khttp.RequestFromServerContext(ctx); ok && request != nil && request.URL != nil {
				if _, allowed := public[Route{Method: request.Method, Path: request.URL.Path}]; allowed {
					return next(ctx, req)
				}
			}
			return protected(ctx, req)
		}
	}
}
