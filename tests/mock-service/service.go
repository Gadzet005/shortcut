package main

import (
	"github.com/Gadzet005/shortcut/pkg/app/di"
	"github.com/Gadzet005/shortcut/pkg/app/lifecycle"
	"github.com/Gadzet005/shortcut/pkg/shortcut"
	"github.com/Gadzet005/shortcut/tests/mock-service/handlers/badresponse"
	"github.com/Gadzet005/shortcut/tests/mock-service/handlers/catalog"
	"github.com/Gadzet005/shortcut/tests/mock-service/handlers/checkout"
	"github.com/Gadzet005/shortcut/tests/mock-service/handlers/dashboard"
	"github.com/Gadzet005/shortcut/tests/mock-service/handlers/orders"
	"github.com/Gadzet005/shortcut/tests/mock-service/handlers/pipeline"
	"github.com/Gadzet005/shortcut/tests/mock-service/handlers/retrytest"
	"github.com/Gadzet005/shortcut/tests/mock-service/handlers/store"
	"github.com/gin-gonic/gin"
)

func newService() lifecycle.App {
	s := &service{}
	s.Container = di.NewContainer[Config](s)
	return di.NewApp(s.Container)
}

type service struct {
	*di.Container[Config]
}

func (s *service) Name() string {
	return "mock-service"
}

func (s *service) Run(ctx lifecycle.Context) error {
	r := s.HTTP("mock-service")
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	{
		g := r.Group("/orders")
		g.POST("/get-users-by-ids", shortcut.New(orders.GetUsersByIDs, s.Logger()))
		g.POST("/get-top-orders", shortcut.New(orders.GetTopOrders, s.Logger()))
		g.POST("/merge-orders-and-users", shortcut.New(orders.MergeOrdersAndUsers, s.Logger()))
	}

	{
		g := r.Group("/test")
		g.POST("/echo-error", shortcut.New(badresponse.EchoError, s.Logger()))
		g.POST("/invalid-content-type", badresponse.InvalidContentType)
		g.POST("/missing-http-response", shortcut.New(badresponse.MissingHTTPResponse, s.Logger()))
		g.POST("/slow-response", badresponse.SlowResponse)
	}

	{
		g := r.Group("/catalog")
		g.POST("/fetch-product", shortcut.New(catalog.FetchProduct, s.Logger()))
		g.POST("/fetch-inventory", shortcut.New(catalog.FetchInventory, s.Logger()))
		g.POST("/fetch-pricing", shortcut.New(catalog.FetchPricing, s.Logger()))
		g.POST("/build-detail", shortcut.New(catalog.BuildDetail, s.Logger()))
	}

	{
		g := r.Group("/dashboard")
		g.POST("/fetch-weather", shortcut.New(dashboard.FetchWeather, s.Logger()))
		g.POST("/fetch-traffic", shortcut.New(dashboard.FetchTraffic, s.Logger()))
		g.POST("/fetch-events", shortcut.New(dashboard.FetchEvents, s.Logger()))
		g.POST("/aggregate-report", shortcut.New(dashboard.AggregateReport, s.Logger()))
	}

	{
		g := r.Group("/checkout")
		g.POST("/parse-request", shortcut.New(checkout.ParseRequest, s.Logger()))
		g.POST("/validate-user", shortcut.New(checkout.ValidateUser, s.Logger()))
		g.POST("/fetch-product", shortcut.New(checkout.FetchProduct, s.Logger()))
		g.POST("/check-inventory", shortcut.New(checkout.CheckInventory, s.Logger()))
		g.POST("/apply-discount", shortcut.New(checkout.ApplyDiscount, s.Logger()))
		g.POST("/build-summary", shortcut.New(checkout.BuildSummary, s.Logger()))
	}

	{
		g := r.Group("/pipeline")
		g.POST("/step1", shortcut.New(pipeline.Step1, s.Logger()))
		g.POST("/step2", shortcut.New(pipeline.Step2, s.Logger()))
		g.POST("/step3", shortcut.New(pipeline.Step3, s.Logger()))
		g.POST("/step4", shortcut.New(pipeline.Step4, s.Logger()))
		g.POST("/step5", shortcut.New(pipeline.Step5, s.Logger()))
	}

	{
		g := r.Group("/retry-test")
		g.POST("/flaky", shortcut.New(retrytest.FlakyEndpoint, s.Logger()))
	}

	{
		g := r.Group("/timeout-test")
		g.POST("/slow-step1", shortcut.New(badresponse.SlowStep1, s.Logger()))
		g.POST("/slow-step2", shortcut.New(badresponse.SlowStep2, s.Logger()))
	}

	{
		g := r.Group("/store")
		g.POST("/validate-item", shortcut.New(store.ValidateItem, s.Logger()))
		g.POST("/save-item", shortcut.New(store.SaveItem, s.Logger()))
		g.POST("/get-all-items", shortcut.New(store.GetAllItems, s.Logger()))
		g.POST("/delete-all-items", shortcut.New(store.DeleteAllItems, s.Logger()))
	}

	return nil
}
