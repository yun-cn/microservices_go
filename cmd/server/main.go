package main

import (
	"fmt"
	lonev1 "github.com/yun-cn/lone-engine/api/proto/loan/v1"
	"github.com/yun-cn/lone-engine/initernal/logger"
	"github.com/yun-cn/lone-engine/initernal/repository"
	"github.com/yun-cn/lone-engine/server"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

const (
	defaultPort       = "50051"
	defaultDBHost     = "localhost"
	defaultDBPort     = "5432"
	defaultDBName     = "loan_engine"
	defaultDBUser     = "loan_engine_user"
	defaultDBPassword = "loan_engine_password"

	// Connection pool defaults
	defaultDBMaxIdleConns    = 10
	defaultDBMaxOpenConns    = 100
	defaultDBConnMaxLifetime = 3600 // seconds

)

func main() {
	log, err := logger.NewLogger(getEnv("ENV", "development"))
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = log.Sync() }()

	log.Info("Starting server")

	db := connectDB(log)

	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	sqlDB.SetMaxIdleConns(getEnvInt("DB_MAX_IDLE_CONNS", defaultDBMaxIdleConns))
	sqlDB.SetMaxOpenConns(getEnvInt("DB_MAX_OPEN_CONNS", defaultDBMaxOpenConns))
	sqlDB.SetConnMaxLifetime(time.Duration(getEnvInt("DB_CONN_MAX_LIFETIME", defaultDBConnMaxLifetime)) * time.Second)

	// Start Server
	grpcServer := buildGRPCServer(db, log)

	address := fmt.Sprintf("0.0.0.0:%s", getEnv("PORT", defaultPort))
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal("Failed to listen", zap.String("address", address), zap.Error(err))
	}

	log.Info("gRPC server listening", zap.String("address", address))

	go handleShutdown(grpcServer, log)

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatal("Failed to serve", zap.Error(err))
	}

}

func connectDB(log *logger.Logger) *gorm.DB {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		getEnv("DB_HOST", defaultDBHost),
		getEnv("DB_USER", defaultDBUser),
		getEnv("DB_PASSWORD", defaultDBPassword),
		getEnv("DB_NAME", defaultDBName),
		getEnv("DB_PORT", defaultDBPort),
	)

	log.Info("Connecting to database",
		zap.String("host", getEnv("DB_HOST", defaultDBHost)),
		zap.String("database", getEnv("DB_NAME", defaultDBName)),
		zap.String("user", getEnv("DB_USER", defaultDBUser)),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		log.Fatal("Failed to connect to database", zap.Error(err))
	}

	log.Info("Connected to database")
	return db
}

func buildGRPCServer(db *gorm.DB, log *logger.Logger) *grpc.Server {
	grpcServer := grpc.NewServer()

	lonev1.RegisterLoanCalculatorServer(grpcServer, server.NewLoanCalculatorServer(
		repository.NewCalculationRecordRepository(db, log.Logger),
		log,
	))

	healthSrv := health.NewServer()
	healthSrv.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(grpcServer, healthSrv)

	reflection.Register(grpcServer)

	return grpcServer
}

func handleShutdown(grpcServer *grpc.Server, log *logger.Logger) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	sig := <-sigChan

	log.Info("Shutting down", zap.String("signal", sig.String()))
	grpcServer.GracefulStop()
	log.Info("Server stopped")
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value != "" {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return defaultValue
}
