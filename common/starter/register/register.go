package register

import (
	"fmt"
	"sync"

	"tanghu.com/go-micro/common/starter/register/eureka"

	"github.com/micro/go-micro/registry"
	"github.com/micro/go-micro/registry/etcd"
	"github.com/spf13/viper"
)

var (
	initOnce sync.Once
	regis    registry.Registry
)

// GetRegistry singleton Registry
func GetRegistry() registry.Registry {
	initOnce.Do(func() {
		var err error
		regis, err = newRegistry()
		if err != nil {
			panic(err)
		}
	})
	if regis == nil {
		panic("regis is nil")
	}

	return regis
}

func newRegistry() (reg registry.Registry, err error) {
	var defaultRegis = viper.GetString("registry.default")
	if defaultRegis == "" {
		defaultRegis = "eureka"
	}

	addressesKey := fmt.Sprintf("registry.%s.addresses", defaultRegis)
	var addresses = viper.GetStringSlice(addressesKey)
	if len(addresses) == 0 {
		return nil, fmt.Errorf("Invalid registry addresses for %s ", defaultRegis)
	}

	switch defaultRegis {
	case "etcd":
		reg = etcd.NewRegistry(registry.Addrs(addresses...))
	case "eureka":
		reg = eureka.NewRegistry(registry.Addrs(addresses...))
	default:
		return nil, fmt.Errorf("not support %s registry", defaultRegis)
	}

	return reg, nil
}
