package initialization

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

const projectDirName = "api-devices"

// InitEnv loads the .env file and prints the environment configuration.
func InitEnv(logger *zap.SugaredLogger) error {
	// Load .env file and print variables
	envFile, err := readEnv()
	logger.Debugf("InitLogger - envFile = %s", envFile)
	if err != nil {
		return fmt.Errorf("failed to load the env file at ./%s: %w", envFile, err)
	}
	printEnv(logger)
	return nil
}

func readEnv() (string, error) {
	// solution taken from https://stackoverflow.com/a/68347834/3590376
	projectName := regexp.MustCompile(`^(.*` + projectDirName + `)`)
	currentWorkDirectory, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("cannot get current working directory: %w", err)
	}
	rootPath := projectName.Find([]byte(currentWorkDirectory))
	envFilePath := filepath.Join(string(rootPath), ".env")
	err = godotenv.Load(envFilePath)
	return envFilePath, err
}

func printEnv(logger *zap.SugaredLogger) {
	logger.Infof("ENVIRONMENT = %s", os.Getenv("ENV"))
	logger.Infof("LOG_FOLDER = %s", os.Getenv("LOG_FOLDER"))
	logger.Info("MONGODB_URL = ****")
	logger.Infof("MQTT_URL = %s", os.Getenv("MQTT_URL"))
	logger.Infof("MQTT_PORT = %s", os.Getenv("MQTT_PORT"))
	logger.Infof("MQTT_TLS = %s", os.Getenv("MQTT_TLS"))
	logger.Infof("MQTT_CA_FILE = %s", os.Getenv("MQTT_CA_FILE"))
	logger.Infof("MQTT_CERT_FILE = %s", os.Getenv("MQTT_CERT_FILE"))
	logger.Infof("MQTT_KEY_FILE = %s", os.Getenv("MQTT_KEY_FILE"))
	logger.Infof("MQTT_CLIENT_ID = %s", os.Getenv("MQTT_CLIENT_ID"))
	logger.Infof("MQTT_AUTH = %s", os.Getenv("MQTT_AUTH"))
	logger.Info("MQTT_USER = ****")
	logger.Info("MQTT_PASSWORD = ****")
	logger.Infof("GRPC_URL = %s", os.Getenv("GRPC_URL"))
	logger.Infof("GRPC_TLS = %s", os.Getenv("GRPC_TLS"))
	logger.Infof("CERT_FOLDER_PATH = %s", os.Getenv("CERT_FOLDER_PATH"))
}
