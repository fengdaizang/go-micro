package eureka

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/hudl/fargo"
	"github.com/micro/go-micro/registry"
	"github.com/spf13/viper"
)

func appToService(app *fargo.Application) []*registry.Service {
	serviceMap := make(map[string]*registry.Service)

	for _, instance := range app.Instances {
		id := instance.Id()
		addr := instance.IPAddr
		port := instance.Port

		var version string
		var metadata map[string]string
		var endpoints []*registry.Endpoint

		// get version
		if _, err := instance.Metadata.GetString("version"); err != nil {
			continue
		}

		if k, err := instance.Metadata.GetString("endpoints"); err == nil {
			json.Unmarshal([]byte(k), &endpoints)
		}

		if k, err := instance.Metadata.GetString("metadata"); err == nil {
			json.Unmarshal([]byte(k), &metadata)
		}

		// get existing service
		service, ok := serviceMap[version]
		if !ok {
			// create new if doesn't exist
			service = &registry.Service{
				Name:      strings.ToLower(app.Name),
				Version:   version,
				Endpoints: endpoints,
			}
		}

		// append node
		service.Nodes = append(service.Nodes, &registry.Node{
			Id:       id,
			Address:  fmt.Sprintf("%s:%d", addr, port),
			Metadata: metadata,
		})

		// save
		serviceMap[version] = service
	}

	services := make([]*registry.Service, 0, len(serviceMap))
	for _, service := range serviceMap {
		services = append(services, service)
	}

	return services
}

// only parses first node
func serviceToInstance(service *registry.Service) (*fargo.Instance, error) {
	if len(service.Nodes) == 0 {
		return nil, errors.New("Require nodes")
	}

	node := service.Nodes[0]
	_, pt, _ := net.SplitHostPort(node.Address)
	port, _ := strconv.Atoi(pt)

	ipAddress := viper.GetString("registry.eureka.ipaddress")

	instance := &fargo.Instance{
		InstanceId:       fmt.Sprintf("%s:%d", ipAddress, port),
		App:              service.Name,
		HostName:         ipAddress,
		IPAddr:           ipAddress,
		VipAddress:       service.Name,
		SecureVipAddress: service.Name,
		Port:             port,
		Status:           fargo.UP,
		CountryId:        1,
		UniqueID: func(i fargo.Instance) string {
			return fmt.Sprintf("%s:%s", node.Address, node.Id)
		},
		DataCenterInfo: fargo.DataCenterInfo{Name: fargo.MyOwn},
	}

	// set version
	instance.SetMetadataString("version", service.Version)

	// set instance ID
	// instance.InstanceId = instance.UniqueID(*instance)
	instance.SetMetadataString("instanceId", node.Id)

	// set endpoints
	if b, err := json.Marshal(service.Endpoints); err == nil {
		instance.SetMetadataString("endpoints", string(b))
	}

	// set metadata
	if b, err := json.Marshal(node.Metadata); err == nil {
		instance.SetMetadataString("metadata", string(b))
	}

	return instance, nil
}
