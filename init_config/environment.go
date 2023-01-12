package init_config

import (
	"fmt"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"os"
	"regexp"
)

const projectDirName = "api-devices"

func InitEnv() (string, error) {
	// solution taken from https://stackoverflow.com/a/68347834/3590376
	projectName := regexp.MustCompile(`^(.*` + projectDirName + `)`)
	currentWorkDirectory, _ := os.Getwd()
	rootPath := projectName.Find([]byte(currentWorkDirectory))
	envFilePath := string(rootPath) + `/.env`
	err := godotenv.Load(envFilePath)
	return envFilePath, err
}

func PrintEnv(logger *zap.SugaredLogger) {
	fmt.Println("ENVIRONMENT = " + os.Getenv("ENV"))
	fmt.Println("MONGODB_URL = " + os.Getenv("MONGODB_URL"))
	fmt.Println("MQTT_URL = " + os.Getenv("MQTT_URL"))
	fmt.Println("MQTT_PORT = " + os.Getenv("MQTT_PORT"))
	fmt.Println("MQTT_TLS = " + os.Getenv("MQTT_TLS"))
	fmt.Println("MQTT_CA_FILE = " + os.Getenv("MQTT_CA_FILE"))
	fmt.Println("MQTT_CERT_FILE = " + os.Getenv("MQTT_CERT_FILE"))
	fmt.Println("MQTT_KEY_FILE = " + os.Getenv("MQTT_KEY_FILE"))
	fmt.Println("MQTT_CLIENT_ID = " + os.Getenv("MQTT_CLIENT_ID"))
	fmt.Println("MQTT_AUTH = " + os.Getenv("MQTT_AUTH"))
	fmt.Println("MQTT_USER = " + os.Getenv("MQTT_USER"))
	fmt.Println("MQTT_PASSWORD = " + os.Getenv("MQTT_PASSWORD"))
	fmt.Println("GRPC_URL = " + os.Getenv("GRPC_URL"))
	fmt.Println("GRPC_TLS = " + os.Getenv("GRPC_TLS"))
	fmt.Println("CERT_FOLDER_PATH = " + os.Getenv("CERT_FOLDER_PATH"))

	logger.Info("ENVIRONMENT = " + os.Getenv("ENV"))
	logger.Info("MONGODB_URL = " + os.Getenv("MONGODB_URL"))
	logger.Info("MQTT_URL = " + os.Getenv("MQTT_URL"))
	logger.Info("MQTT_PORT = " + os.Getenv("MQTT_PORT"))
	logger.Info("MQTT_TLS = " + os.Getenv("MQTT_TLS"))
	logger.Info("MQTT_CA_FILE = " + os.Getenv("MQTT_CA_FILE"))
	logger.Info("MQTT_CERT_FILE = " + os.Getenv("MQTT_CERT_FILE"))
	logger.Info("MQTT_KEY_FILE = " + os.Getenv("MQTT_KEY_FILE"))
	logger.Info("MQTT_CLIENT_ID = " + os.Getenv("MQTT_CLIENT_ID"))
	logger.Info("MQTT_AUTH = " + os.Getenv("MQTT_AUTH"))
	logger.Info("MQTT_USER = " + os.Getenv("MQTT_USER"))
	logger.Info("MQTT_PASSWORD = " + os.Getenv("MQTT_PASSWORD"))
	logger.Info("GRPC_URL = " + os.Getenv("GRPC_URL"))
	logger.Info("GRPC_TLS = " + os.Getenv("GRPC_TLS"))
	logger.Info("CERT_FOLDER_PATH = " + os.Getenv("CERT_FOLDER_PATH"))
}
