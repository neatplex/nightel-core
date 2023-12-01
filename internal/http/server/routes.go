package server

import (
	"github.com/labstack/echo/v4/middleware"
	"github.com/neatplex/nightel-core/internal/http/server/handlers"
	v12 "github.com/neatplex/nightel-core/internal/http/server/handlers/v1"
	mw "github.com/neatplex/nightel-core/internal/http/server/middleware"
)

func (s *Server) registerRoutes() {
	s.E.GET("/healthz", handlers.Healthz)

	v1Api := s.E.Group("/api/v1", middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(5)))
	{
		public := v1Api.Group("", middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(1)))
		{
			public.POST("/auth/sign-up", v12.AuthSignUp(s.container))
			public.POST("/auth/sign-in", v12.AuthSignIn(s.container))
		}

		private := v1Api.Group("", mw.Authorize(s.container))
		{
			private.GET("/stories", v12.StoriesIndex(s.container))
		}
	}
}