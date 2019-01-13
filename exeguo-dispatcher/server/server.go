package server

import (
	"net/http"

	"github.com/fabiosussetto/exeguo/security"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	adapter "github.com/gwatts/gin-adapter"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

type ServerConfig struct {
	ServerAddress string
	PathToDB      string
}

func setupDB(config ServerConfig) *gorm.DB {
	db, err := gorm.Open("sqlite3", config.PathToDB)
	if err != nil {
		panic("failed to connect database")
	}

	db.Exec("PRAGMA foreign_keys = ON;")
	db.LogMode(true)

	db.AutoMigrate(&Config{}, &TargetHost{}, &ExecutionPlan{}, &ExecutionPlanHost{}, &ExecutionPlanRun{}, &RunStatus{}, &User{})

	var (
		tlsCaKey          Config
		tlsCaCert         Config
		tlsDispatcherKey  Config
		tlsDispatcherCert Config
	)

	caCertData, err := security.GenerateRootCertAndKey()
	dispatcherCertData, err := security.GenerateClientCertAndKey(caCertData)

	if db.Where(&Config{Key: "tls.ca_key"}).First(&tlsCaKey).RecordNotFound() {
		db.Create(&Config{Key: "tls.ca_key", Value: string(caCertData.PrivateKeyPEM)})
	}

	if db.Where(&Config{Key: "tls.ca_cert"}).First(&tlsCaCert).RecordNotFound() {
		db.Create(&Config{Key: "tls.ca_cert", Value: string(caCertData.CertPEM)})
	}

	if db.Where(&Config{Key: "tls.dispatcher_key"}).First(&tlsDispatcherKey).RecordNotFound() {
		db.Create(&Config{Key: "tls.dispatcher_key", Value: string(dispatcherCertData.PrivateKeyPEM)})
	}

	if db.Where(&Config{Key: "tls.dispatcher_cert"}).First(&tlsDispatcherCert).RecordNotFound() {
		db.Create(&Config{Key: "tls.dispatcher_cert", Value: string(dispatcherCertData.CertPEM)})
	}

	return db
}

func StartServer(config ServerConfig) {
	db := setupDB(config)
	defer db.Close()

	ab := SetupAuth(db)

	router := gin.Default()

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"http://localhost:3000"}
	corsConfig.AllowCredentials = true
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD"}

	router.Use(cors.New(corsConfig))
	router.Use(adapter.Wrap(ab.LoadClientStateMiddleware))

	env := &Env{db: db, Auth: ab}

	v1 := router.Group("/v1")
	{
		// v1.Use(adapter.Wrap(authboss.Middleware(ab, true, false, false)))

		hostsRoute := v1.Group("/hosts")
		{
			hostsRoute.GET("/", env.HostListEndpoint)
			hostsRoute.POST("/", env.HostCreateEndpoint)
			hostsRoute.GET("/:id", env.HostDetailEndpoint)
			hostsRoute.PUT("/:id", env.HostUpdateEndpoint)
			hostsRoute.DELETE("/:id", env.HostDeleteEndpoint)
		}

		hostStatusRoute := v1.Group("/host-statuses")
		{
			hostStatusRoute.POST("/", env.HostStatusEndpoint)
		}

		commandRunR := v1.Group("/exec-plans")
		{
			commandRunR.POST("/", env.ExecutionPlanCreateEndpoint)
			commandRunR.GET("/:id", env.ExecutionPlanDetailEndpoint)
		}

		execPlanRunRoute := v1.Group("/exec-plan-runs")
		{
			execPlanRunRoute.POST("/", env.ExecutionPlanRunCreateEndpoint)
			execPlanRunRoute.GET("/:id", env.ExecutionPlanRunDetailEndpoint)
		}

		stopPlanReqRoute := v1.Group("/stop-plan-requests")
		{
			stopPlanReqRoute.POST("/", env.CreateStopPlanRequest)
		}

	}
	router.GET("/user", env.UserDetail)

	router.Any("/auth/*w", gin.WrapH(ab.LoadClientStateMiddleware(http.StripPrefix("/auth", ab.Config.Core.Router))))

	router.StaticFS("/ui", Assets)

	router.Run(config.ServerAddress)
}
