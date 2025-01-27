package discover_legacy

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"reflect"

	"github.com/dillya/melo-webapi/internal/device"
	"github.com/dillya/melo-webapi/internal/utils/middleware"

	"github.com/danielgtaylor/huma/v2"
)

// Device interface
type legacyDeviceInterface struct {
	Address   string `json:"address" example:"192.168.0.100" doc:"The IP address of the network interface"`
	HwAddress string `json:"hw_address" example:"01:23:45:67:89:ab" doc:"The MAC address of the network interface"`
}

// Device
type legacyDevice struct {
	Name   string                  `json:"name" example:"Living room" doc:"Name of the device"`
	Serial string                  `json:"serial" example:"01:23:45:67:89:ab" doc:"Serial Number of the device"`
	Port   uint16                  `json:"port" example:"8080" minimum:"0" maximum:"65535" doc:"HTTP port of the device API"`
	List   []legacyDeviceInterface `json:"list" doc:"List of network interfaces of the device"`
}

// Legacy discover response
type legacyDiscoverOutput struct {
	Body any
}

func createQueryError(query string, value any) error {
	return huma.Error422UnprocessableEntity("invalid action", &huma.ErrorDetail{
		Message:  "required query parameter is missing",
		Location: "query." + query,
		Value:    value,
	})
}

func convertInterface(ifaces []device.DeviceInterface) []legacyDeviceInterface {
	// Convert interface list
	list := []legacyDeviceInterface{}

	// Generate list
	for _, iface := range ifaces {
		// Add interface to list
		list = append(list, legacyDeviceInterface{
			Address:   iface.Ipv4Address,
			HwAddress: iface.MacAddress,
		})
	}

	return list
}

func listDevice(ctx context.Context, db *sql.DB, ip string) []legacyDevice {
	// List devices
	devices := device.List(ctx, db, ip)

	// Convert list to legacy one
	list := []legacyDevice{}
	for _, device := range devices {
		// Add device to list
		list = append(list, legacyDevice{
			Name:   device.Name,
			Serial: device.Serial,
			Port:   device.HttpPort,
			List:   convertInterface(device.Interfaces),
		})
	}

	return list
}

func Register(api huma.API, db *sql.DB) {
	// Register responses to the API (same handler is shared for many kind of responses)
	registry := api.OpenAPI().Components.Schemas
	schema := &huma.Schema{
		OneOf: []*huma.Schema{
			registry.Schema(reflect.TypeOf(legacyDevice{}), true, ""),
		},
	}

	// Register GET /discover handler
	huma.Register(api, huma.Operation{
		OperationID: "legacyDiscover",
		Method:      http.MethodGet,
		Path:        "/discover",
		Summary:     "[Deprecated] Discover device API",
		Description: "List, add and remove devices and their interfaces.",
		Deprecated:  true,
		Middlewares: huma.Middlewares{middleware.GetIpExtractor()},
		Responses: map[string]*huma.Response{
			"200": {
				Content: map[string]*huma.MediaType{
					"application/json": {
						Schema: schema,
					},
				},
			},
		},
	}, func(ctx context.Context, input *struct {
		Action    string `query:"action" example:"list" enum:"list,add_device,remove_device,add_address,remove_address" required:"true"`
		Serial    string `query:"serial" example:"01:23:45:67:89:ab" doc:"The serial number of the device"`
		Name      string `query:"name" example:"Living Room" doc:"The device name when action is 'add_device'"`
		HostName  string `query:"hostname" example:"melo-living-room" doc:"The hostname of the device when action is 'add_device'"`
		HttpPort  uint16 `query:"port" example:"80" doc:"The HTTP port when action is 'add_device'"`
		HttpsPort uint16 `query:"sport" example:"443" doc:"The HTTPs port when action is 'add_device'"`
		HwAddress string `query:"hw_address" example:"01:23:45:67:89:ab" doc:"The Mac address of the interface when action is 'add_address'"`
		Address   string `query:"address" example:"192.168.0.100" doc:"The IP address of the interface when action is 'add_address'"`
	},
	) (*legacyDiscoverOutput, error) {
		resp := &legacyDiscoverOutput{}
		var err error = nil

		// Get IP address of the remote
		ip := middleware.ExtractIp(ctx)

		// Parse the action
		switch input.Action {
		case "list":
			resp.Body = listDevice(ctx, db, ip)
		case "add_device":
			// Check required query
			if input.Serial == "" {
				err = createQueryError("serial", input.Serial)
			} else if input.Name == "" {
				err = createQueryError("name", input.Name)
			} else if input.HttpPort == 0 {
				err = createQueryError("port", input.HttpPort)
			} else if !device.Add(ctx, db, ip, device.Device{Serial: input.Serial, Name: input.Name, HttpPort: input.HttpPort}) {
				err = huma.Error500InternalServerError("failed to add device")
			} else {
				resp.Body = struct{}{}
			}
		case "remove_device":
			// Check required query
			if input.Serial == "" {
				err = createQueryError("serial", input.Serial)
			} else if !device.Remove(ctx, db, ip, input.Serial) {
				err = huma.Error404NotFound("device not found")
			} else {
				resp.Body = struct{}{}
			}
		case "add_address":
			// Check required query
			if input.Serial == "" {
				err = createQueryError("serial", input.Serial)
			} else if input.HwAddress == "" {
				err = createQueryError("hw_address", input.HwAddress)
			} else if input.Address == "" {
				err = createQueryError("address", input.Address)
			} else if !device.AddAddress(ctx, db, ip, input.Serial, device.DeviceInterface{MacAddress: input.HwAddress, Ipv4Address: input.Address}, true) {
				err = huma.Error500InternalServerError("failed to add address")
			} else {
				resp.Body = struct{}{}
			}
		case "remove_address":
			// Check required query
			if input.Serial == "" {
				err = createQueryError("serial", input.Serial)
			} else if input.HwAddress == "" {
				err = createQueryError("hw_address", input.HwAddress)
			} else if !device.RemoveAddress(ctx, db, ip, input.Serial, input.HwAddress, true) {
				err = huma.Error404NotFound("address not found")
			} else {
				resp.Body = struct{}{}
			}
		default:
			fmt.Printf("Invalid action %s\n", input.Action)
			err = huma.Error422UnprocessableEntity("invalid action", fmt.Errorf("Action '%s' not supported", input.Action))
		}
		return resp, err
	})
}
