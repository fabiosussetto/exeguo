package server

import (
	"github.com/fabiosussetto/exeguo/security"
	"github.com/gin-gonic/gin"
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

	db.AutoMigrate(&Config{}, &TargetHost{}, &ExecutionPlan{}, &ExecutionPlanHost{}, &ExecutionPlanRun{}, &RunStatus{})

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

	router := gin.Default()

	env := &Env{db: db}

	v1 := router.Group("/v1")
	{
		commandsR := v1.Group("/hosts")
		{
			commandsR.GET("/", env.HostListEndpoint)
			commandsR.POST("/", env.HostCreateEndpoint)
			commandsR.PUT("/:id", env.HostUpdateEndpoint)
			commandsR.DELETE("/:id", env.HostDeleteEndpoint)
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

	}

	router.Run(config.ServerAddress)
}
