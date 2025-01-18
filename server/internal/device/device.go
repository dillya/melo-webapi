package device

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"github.com/dillya/melo-webapi/internal/utils/middleware"
)

// Device List
type deviceListOutput struct {
	Body []Device
}

// Operation result
type resultOutput struct {
	Body result
}

// Result
type result struct {
	Code   uint   `json:"code" example:"2" doc:"The result code: 0=success"`
	Error  string `json:"error,omitempty" example:"Failed to add device" doc:"The error message if code != 0"`
}

// Interface
type Interface struct {
	Type          string        `json:"type" example:"eth" enum:"unknown,eth,wifi" doc:"The network interface type"`
	Name          string        `json:"name" example:"Unknown" doc:"The name of the interface"`
	MacAddress    string        `json:"mac" example:"01:23:45:67:89:ab" doc:"The MAC address of the network interface"`
	Ipv4Address   string        `json:"ipv4,omitempty" example:"192.168.0.100" doc:"The IPv4 address of the network interface"`
	Ipv6Address   string        `json:"ipv6,omitempty" example:"fe80::5814:a424:50e8:81b0/64" doc:"The IPv6 address of the network interface"`
}

// Device
type Device struct {
	Serial      string            `json:"serial" example:"01:23:45:67:89:ab" doc:"Serial Number of the device"`
	Name        string            `json:"name" example:"Living room" doc:"Name of the device"`
	Description string            `json:"description,omitempty" example:"Melo of Library" doc:"Description of the device"`
	Icon        string            `json:"icon,omitempty" example:"living" enum:"living,kitchen,bed" doc:"Icon to distinguish devices"`
	Location    string            `json:"location,omitempty" example:"Living room library" doc:"The exact location of the device"`
	HttpPort    uint16            `json:"http_port" example:"8080" minimum:"0" maximum:"65535" doc:"HTTP port of the device API"`
	HttpsPort   uint16            `json:"https_port,omitempty" example:"8443" minimum:"0" maximum:"65535" doc:"HTTPs port of the device API"`
	Online      bool              `json:"online" example:"true" doc:"The device online status"`
	LastUpdate  uint64            `json:"last_update" example:"0" doc:"The last update timestamp as Unix epoch (updated on every PUT methods)"`
	Interfaces  []Interface `json:"ifaces" doc:"List of network interfaces of the device"`
}

func Register(api huma.API, db *sql.DB) {
	// IP client extractor middleware
	client_ip_extract := middleware.GetIpExtractor()

	// Register GET /device/list handler
	huma.Register(api, huma.Operation{
		OperationID: "listDevice",
		Method:      http.MethodGet,
		Path:        "/device/list",
		Summary:     "List devices",
		Description: "List all devices registered on the local network.",
		Tags:        []string{"Device"},
		Middlewares: huma.Middlewares{client_ip_extract},
	}, func(ctx context.Context, input *struct{}) (*deviceListOutput, error) {
		ip := middleware.ExtractIp(ctx)

		// List devices
		resp := &deviceListOutput{}
		resp.Body = List(ctx, db, ip)
		return resp, nil
	})

	// Register PUT /device/add handler
	huma.Register(api, huma.Operation{
		OperationID: "addDevice",
		Method:      http.MethodPut,
		Path:        "/device/add",
		Summary:     "Add / reset a device",
		Description: "Add a new device / reset a device on the local network.",
		Tags:        []string{"Device"},
		Middlewares: huma.Middlewares{client_ip_extract},
	}, func(ctx context.Context, input *struct{
		Body Device
	}) (*resultOutput, error) {
		ip := middleware.ExtractIp(ctx)

		// Add device
		resp := &resultOutput{}
		if !Add(ctx, db, ip, input.Body.Serial, input.Body.Name, input.Body.HttpPort) {
			resp.Body.Code = 1
			resp.Body.Error = "Failed to add device"
		}

		return resp, nil
	})

	// Register DELETE /device/{serial} handler
	huma.Register(api, huma.Operation{
		OperationID: "removeDevice",
		Method:      http.MethodDelete,
		Path:        "/device/{serial}",
		Summary:     "Remove the device",
		Description: "Remove the device from the local network.",
		Tags:        []string{"Device"},
		Middlewares: huma.Middlewares{client_ip_extract},
	}, func(ctx context.Context, input *struct{
		Serial string `path:"serial" example:"01:23:45:67:89:ab" doc:"Serial Number of the device to remove"`
	}) (*resultOutput, error) {
		ip := middleware.ExtractIp(ctx)

		// Remove device
		resp := &resultOutput{}
		if !Remove(ctx, db, ip, input.Serial) {
			resp.Body.Code = 1
			resp.Body.Error = "Failed to remove device"
		}

		return resp, nil
	})

	// Register PUT /device/{serial}/online handler
	huma.Register(api, huma.Operation{
		OperationID: "updateDeviceOnlineStatus",
		Method:      http.MethodPut,
		Path:        "/device/{serial}/online",
		Summary:     "Set the device as online",
		Description: "Set the device as online and update timestamp.",
		Tags:        []string{"Device"},
		Middlewares: huma.Middlewares{client_ip_extract},
	}, func(ctx context.Context, input *struct{
		Serial string `path:"serial" example:"01:23:45:67:89:ab" doc:"Serial Number of the device to modify"`
	}) (*resultOutput, error) {
		ip := middleware.ExtractIp(ctx)

		// Set device online
		resp := &resultOutput{}
		if !UpdateStatus(ctx, db, ip, input.Serial, true) {
			resp.Body.Code = 1
			resp.Body.Error = "Failed to set device online"
		}

		return resp, nil
	})

	// Register PUT /device/{serial}/offline handler
	huma.Register(api, huma.Operation{
		OperationID: "updateDeviceOfflineStatus",
		Method:      http.MethodPut,
		Path:        "/device/{serial}/offline",
		Summary:     "Set the device as offline",
		Description: "Set the device as offline and update timestamp.",
		Tags:        []string{"Device"},
		Middlewares: huma.Middlewares{client_ip_extract},
	}, func(ctx context.Context, input *struct{
		Serial string `path:"serial" example:"01:23:45:67:89:ab" doc:"Serial Number of the device to modify"`
	}) (*resultOutput, error) {
		ip := middleware.ExtractIp(ctx)

		// Set device offline
		resp := &resultOutput{}
		if !UpdateStatus(ctx, db, ip, input.Serial, false) {
			resp.Body.Code = 1
			resp.Body.Error = "Failed to set device offline"
		}

		return resp, nil
	})

	// Register PUT /device/{serial}/add handler
	huma.Register(api, huma.Operation{
		OperationID: "addNetworkInterface",
		Method:      http.MethodPut,
		Path:        "/device/{serial}/add",
		Summary:     "Add / update a network interface",
		Description: "Add / update a network interface of the device.",
		Tags:        []string{"Device"},
		Middlewares: huma.Middlewares{client_ip_extract},
	}, func(ctx context.Context, input *struct{
		Serial string `path:"serial" example:"01:23:45:67:89:ab" doc:"Serial Number of the device to modify"`
		Body Interface
	}) (*resultOutput, error) {
		ip := middleware.ExtractIp(ctx)

		// Add interface
		resp := &resultOutput{}
		if !AddAddress(ctx, db, ip, input.Serial, input.Body.MacAddress, input.Body.Ipv4Address) {
			resp.Body.Code = 1
			resp.Body.Error = "Failed to add interface"
		}

		return resp, nil
	})

	// Register DELETE /device/{serial}/{mac} handler
	huma.Register(api, huma.Operation{
		OperationID: "removeNetworkInterface",
		Method:      http.MethodDelete,
		Path:        "/device/{serial}/{mac}",
		Summary:     "Remove the network interface",
		Description: "Remove the network interface from the device.",
		Tags:        []string{"Device"},
		Middlewares: huma.Middlewares{client_ip_extract},
	}, func(ctx context.Context, input *struct{
		Serial string `path:"serial" example:"01:23:45:67:89:ab" doc:"Serial Number of the device to modify"`
		Mac string `path:"mac" example:"01:23:45:67:89:ab" doc:"The MAC address of the network interface"`
	}) (*resultOutput, error) {
		ip := middleware.ExtractIp(ctx)

		// Remove interface
		resp := &resultOutput{}
		if !RemoveAddress(ctx, db, ip, input.Serial, input.Mac) {
			resp.Body.Code = 1
			resp.Body.Error = "Failed to remove interface"
		}

		return resp, nil
	})

}
