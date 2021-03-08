package starter

import (
	"fmt"
	"strings"
	"sync"

	"github.com/micro/go-micro/registry"
	"github.com/micro/go-micro/web"
	"github.com/spf13/viper"

	"tanghu.com/go-micro/common/config"
	"tanghu.com/go-micro/common/constant"
	"tanghu.com/go-micro/common/log"
	"tanghu.com/go-micro/common/starter/register"
)

const (
	configFileName = "config"
)

var (
	theWebService     web.Service
	theWebServiceOnce sync.Once
)

// StartAPIServer creates a micro-web-service server and initilize, start it
func StartAPIServer(name, version string) web.Service {
	err := initConfig(name)
	if err != nil {
		panic(err)
	}

	err = log.Init(&log.Options{})
	if err != nil {
		panic(err)
	}

	service := createAPIService(name, version)
	service.Init()

	return service
}

func initConfig(name string) error {
	envVarPrefix := constant.ServiceEnvVarPrefix + strings.ToUpper(name)
	configOpts := &config.Options{
		EnvVarEnabled: true,
		EnvVarPrefix:  envVarPrefix,

		ConfigFileName: configFileName,
	}

	return config.Init(configOpts)
}

func createAPIService(name, version string) web.Service {
	theWebServiceOnce.Do(func() {
		var regis registry.Registry
		// Generated the service name based the WebServiceNameFormat
		serviceName := generateServiceName(name, version)

		enable := viper.GetBool("registry.enable")
		if enable {
			regis = register.GetRegistry()
		}

		serviceOpts := []web.Option{
			web.Name(serviceName),
			web.Address(getServiceListenAddr()),
			// web.WrapHandler(handler.MiddlewareList()...),
			web.Registry(regis),
		}

		service := web.NewService(serviceOpts...)
		service.Init()
		theWebService = service
	})

	return theWebService
}

func generateServiceName(name, version string) string {
	return fmt.Sprintf(constant.MicroServiceNameFormat, name, version)
}

func getServiceListenAddr() string {
	svrListenAddr := viper.GetString("service.server.listenAddr")
	if svrListenAddr == "" {
		svrListenAddr = ":9080"
	}

	return svrListenAddr
}
