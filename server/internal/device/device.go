package device

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

// Device List
type DeviceListOutput struct {
	Body []Device
}

// Device interface
type DeviceInterface struct {
	Address   string `json:"address" example:"192.168.0.100" doc:"The IP address of the network interface"`
	HwAddress string `json:"hw_address" example:"01:23:45:67:89:ab" doc:"The MAC address of the network interface"`
}

// Device
type Device struct {
	Name       string            `json:"name" example:"Living room" doc:"Name of the device"`
	Serial     string            `json:"serial" example:"01:23:45:67:89:ab" doc:"Serial Number of the device"`
	Port       int               `json:"port" example:"8080" doc:"HTTP port of the device API"`
	Interfaces []DeviceInterface `json:"ifaces" doc:"List of network interfaces of the device"`
}

func Register(api huma.API) {
	// Register GET /device/list handler.
	huma.Register(api, huma.Operation{
		OperationID: "listDevice",
		Method:      http.MethodGet,
		Path:        "/device/list",
		Summary:     "List devices",
		Description: "List all devices registered on the local network.",
		Tags:        []string{"Device"},
	}, func(ctx context.Context, input *struct{}) (*DeviceListOutput, error) {
		resp := &DeviceListOutput{}
		resp.Body = []Device{}
		return resp, nil
	})
}
