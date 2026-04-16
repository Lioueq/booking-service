package app

import (
	"booking/internal/container"
	middleware "booking/internal/handlers"
	authhandler "booking/internal/handlers/auth"
	bookinghandler "booking/internal/handlers/booking"
	"booking/internal/handlers/info"
	roomhandler "booking/internal/handlers/room"
	schedulehandler "booking/internal/handlers/schedule"
	slothandler "booking/internal/handlers/slot"
	"booking/internal/logger"
	authusecase "booking/internal/usecase/auth"
	bookingusecase "booking/internal/usecase/booking"
	roomusecase "booking/internal/usecase/room"
	scheduleusecase "booking/internal/usecase/schedule"
	slotusecase "booking/internal/usecase/slot"
	"booking/internal/utils"
	jwtservice "booking/internal/utils/jwtservice"
	"context"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type App struct {
	container *container.Container
	router    *gin.Engine
	address   string
}

func NewApp(c *container.Container) *App {
	router := gin.New()
	router.Use(gin.Recovery(), gin.Logger())
	setUpRoutes(c, router)

	return &App{
		container: c,
		router:    router,
		address:   c.Config.Address,
	}
}

func (a *App) Run(ctx context.Context) error {
	srv := &http.Server{
		Addr:    a.address,
		Handler: a.router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.container.Logger.Error("server error", logger.Err(err))
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		a.container.Logger.Error("server shutdown error", logger.Err(err))
		return err
	}

	a.container.Logger.Info("server shutdown gracefully")
	return nil
}

func setUpRoutes(container *container.Container, router *gin.Engine) {
	jwtSvc := jwtservice.New(os.Getenv("JWT_SECRET"), 24*time.Hour)
	authUC := authusecase.New(container.Storage, jwtSvc)
	authH := authhandler.New(container.Logger, authUC)
	roomUC := roomusecase.New(container.Storage)
	roomH := roomhandler.New(container.Logger, roomUC)
	scheduleUC := scheduleusecase.New(container.Storage)
	scheduleH := schedulehandler.New(container.Logger, scheduleUC)
	slotUC := slotusecase.New(container.Storage)
	slotH := slothandler.New(container.Logger, slotUC)
	bookingUC := bookingusecase.New(container.Storage)
	bookingH := bookinghandler.New(container.Logger, bookingUC)

	router.Static("/docs", "./docs")
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler,
		ginSwagger.URL("/docs/api.yaml"),
	))
	router.GET("/_info", info.Handle())
	router.POST("/register", authH.Register())
	router.POST("/login", authH.Login())
	router.POST("/dummyLogin", authH.DummyLogin())

	protected := router.Group("/")
	protected.Use(middleware.RequireAuth(jwtSvc))

	roomsGroup := protected.Group("/rooms")
	roomsGroup.POST("/create", middleware.RequireRole(utils.RoleAdmin), roomH.CreateRoom())
	roomsGroup.GET("/list", roomH.GetRooms())
	roomsGroup.POST("/:roomId/schedule/create", middleware.RequireRole(utils.RoleAdmin), scheduleH.CreateSchedule())
	roomsGroup.GET("/:roomId/slots/list", slotH.ListSlots())

	bookingsGroup := protected.Group("/bookings")
	bookingsGroup.POST("/create", middleware.RequireRole(utils.RoleUser), bookingH.CreateBooking())
	bookingsGroup.POST("/:bookingId/cancel", middleware.RequireRole(utils.RoleUser), bookingH.CancelBooking())
	bookingsGroup.GET("/my", middleware.RequireRole(utils.RoleUser), bookingH.MyBookings())
	bookingsGroup.GET("/list", middleware.RequireRole(utils.RoleAdmin), bookingH.ListBookings())
}
