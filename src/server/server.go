package server

import (
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/PayCryps/WatchdogGo/src/graph"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/vektah/gqlparser/v2/ast"
)

func HomeHandler(c *gin.Context, startTime time.Time) {
	uptime := time.Since(startTime)

	c.JSON(200, gin.H{
		"now":    time.Now(),
		"uptime": uptime.String(),
		"health": "ok",
	})
}

func graphqlHandler() gin.HandlerFunc {
	srv := handler.New(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{}}))

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	return func(c *gin.Context) {
		srv.ServeHTTP(c.Writer, c.Request)
	}
}

func playgroundHandler() gin.HandlerFunc {
	h := playground.Handler("GraphQL", "/graphql/query")

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

func RegisterRoutes(r *gin.Engine, logger zerolog.Logger) {

	startTime := time.Now()

	r.Static("/static", "./static")

	r.POST("/graphql/query", graphqlHandler())
	r.GET("/graphql", playgroundHandler())

	r.Use(GinContextToContextMiddleware())

	r.GET("/", func(c *gin.Context) {
		HomeHandler(c, startTime)
	})
}
