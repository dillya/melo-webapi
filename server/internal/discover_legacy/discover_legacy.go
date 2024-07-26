package discover_legacy

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	log "github.com/sirupsen/logrus"
)

// Device interface
type LegacyDeviceInterface struct {
	Address   string `json:"address" example:"192.168.0.100" doc:"The IP address of the network interface"`
	HwAddress string `json:"hw_address" example:"01:23:45:67:89:ab" doc:"The MAC address of the network interface"`
}

// Device
type LegacyDevice struct {
	Name   string                  `json:"name" example:"Living room" doc:"Name of the device"`
	Serial string                  `json:"serial" example:"01:23:45:67:89:ab" doc:"Serial Number of the device"`
	Port   int                     `json:"port" example:"8080" doc:"HTTP port of the device API"`
	List   []LegacyDeviceInterface `json:"list" doc:"List of network interfaces of the device"`
}

// Legacy discover response
type LegacyDiscoverOutput struct {
	Body any
}

func createQueryError(query string, value any) error {
	return huma.Error422UnprocessableEntity("invalid action", &huma.ErrorDetail{
		Message:  "required query parameter is missing",
		Location: "query." + query,
		Value:    value,
	})
}

func uint64FromHwAddress(address string) uint64 {
	value := uint64(0)
	parts := strings.Split(address, ":")
	for i, part := range parts {
		v, err := strconv.ParseUint(part, 16, 8)
		if err == nil {
			value |= v << (8 * (5 - i))
		}
	}
	return value
}

func uint64ToHwAddress(address uint64) string {
	value := ""
	for i := 5; i >= 0; i-- {
		v := (address >> (8 * i)) & 0xff
		value += strconv.FormatUint(v, 16) + ":"
	}
	return value[:len(value)-1]
}

func listDeviceInterface(ctx context.Context, db *sql.DB, id int) []LegacyDeviceInterface {
	// Create interface list
	list := []LegacyDeviceInterface{}

	// Fetch interfaces of the current device
	ifaces, err := db.QueryContext(ctx, "SELECT mac, INET_NTOA(ip) FROM device_iface WHERE device_id=?", id)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("failed to get interface list")
		return list
	}
	defer ifaces.Close()

	// Generate list
	for ifaces.Next() {
		// Scan interface
		var mac uint64
		var ip string
		if err := ifaces.Scan(&mac, &ip); err != nil {
			log.WithFields(log.Fields{"error": err}).Error("failed to scan interface")
			continue
		}

		// Add interface to list
		list = append(list, LegacyDeviceInterface{
			Address:   ip,
			HwAddress: uint64ToHwAddress(mac),
		})
	}

	return list
}

func listDevice(ctx context.Context, db *sql.DB, ip string) []LegacyDevice {
	// Create device list
	list := []LegacyDevice{}

	// Fetch devices
	devices, err := db.QueryContext(ctx, "SELECT id, name, serial, http_port FROM device WHERE ip=INET_ATON(?)", ip)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("failed to get device list")
		return list
	}
	defer devices.Close()

	// Generate list
	for devices.Next() {
		// Scan device
		var id, port int
		var serial, name string
		if err := devices.Scan(&id, &name, &serial, &port); err != nil {
			log.WithFields(log.Fields{"error": err}).Error("failed to scan device")
			continue
		}

		// Add device to list
		list = append(list, LegacyDevice{
			Name:   name,
			Serial: serial,
			Port:   port,
			List:   listDeviceInterface(ctx, db, id),
		})
	}

	return list
}

func addDevice(ctx context.Context, db *sql.DB, ip string, serial string, name string, port uint16) bool {
	// Add or update device
	result, err := db.ExecContext(ctx, "INSERT INTO device (serial, ip, name, http_port) VALUES (?, INET_ATON(?), ?, ?) ON DUPLICATE KEY UPDATE name=?, http_port=?",
		serial,
		ip,
		name,
		port,
		name,
		port,
	)
	if err != nil {
		log.WithFields(log.Fields{"error": err, "serial": serial}).Error("failed to add device")
		return false
	}
	_, err = result.RowsAffected()
	return err == nil
}

func removeDevice(ctx context.Context, db *sql.DB, ip string, serial string) bool {
	// Remove device (interfaces will be removed automatically)
	result, err := db.ExecContext(ctx, "DELETE FROM device WHERE ip=INET_ATON(?) AND serial=?",
		ip,
		serial,
	)
	if err != nil {
		log.WithFields(log.Fields{"error": err, "serial": serial}).Error("failed to remove device")
		return false
	}
	rows, err := result.RowsAffected()
	return err == nil && rows == 1
}

func addAddress(ctx context.Context, db *sql.DB, ip string, serial string, hw_address string, address string) bool {
	// Add or update address
	result, err := db.ExecContext(ctx, "INSERT INTO device_iface (device_id, mac, ip) SELECT id, ?, INET_ATON(?) FROM device WHERE ip=INET_ATON(?) AND serial=? ON DUPLICATE KEY UPDATE ip=INET_ATON(?)",
		uint64FromHwAddress(hw_address),
		address,
		ip,
		serial,
		address,
	)
	if err != nil {
		log.WithFields(log.Fields{"error": err, "serial": serial}).Error("failed to add address")
		return false
	}
	_, err = result.RowsAffected()
	return err == nil
}

func removeAddress(ctx context.Context, db *sql.DB, ip string, serial string, hw_address string) bool {
	// Remove address
	result, err := db.ExecContext(ctx, "DELETE FROM device_iface WHERE device_id IN (SELECT id FROM device WHERE ip=INET_ATON(?) AND serial=?) AND mac=?\n",
		ip,
		serial,
		uint64FromHwAddress(hw_address),
	)
	if err != nil {
		log.WithFields(log.Fields{"error": err, "serial": serial}).Error("failed to remove address")
		return false
	}
	rows, err := result.RowsAffected()
	return err == nil && rows == 1
}

func Register(api huma.API, db *sql.DB) {
	// Get proxy IP address header from environment
	ip_header := os.Getenv("MELO_WEBAPI_REAL_IP_HEADER")

	// Create closure for client IP extract
	client_ip_extract := func(ctx huma.Context, next func(huma.Context)) {
		if ip := ctx.Header(ip_header); ip != "" {
			ctx = huma.WithValue(ctx, "remote-ip", ip)
		} else {
			ctx = huma.WithValue(ctx, "remote-ip", strings.Split(ctx.RemoteAddr(), ":")[0])
		}
		next(ctx)
	}

	// Register responses to the API (same handler is shared for many kind of responses)
	registry := api.OpenAPI().Components.Schemas
	schema := &huma.Schema{
		OneOf: []*huma.Schema{
			registry.Schema(reflect.TypeOf(LegacyDevice{}), true, ""),
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
		Middlewares: huma.Middlewares{client_ip_extract},
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
	) (*LegacyDiscoverOutput, error) {
		resp := &LegacyDiscoverOutput{}
		var err error = nil

		// Get IP address of the remote
		ip, _ := ctx.Value("remote-ip").(string)

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
			} else if !addDevice(ctx, db, ip, input.Serial, input.Name, input.HttpPort) {
				err = huma.Error500InternalServerError("failed to add device")
			} else {
				resp.Body = struct{}{}
			}
		case "remove_device":
			// Check required query
			if input.Serial == "" {
				err = createQueryError("serial", input.Serial)
			} else if !removeDevice(ctx, db, ip, input.Serial) {
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
			} else if !addAddress(ctx, db, ip, input.Serial, input.HwAddress, input.Address) {
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
			} else if !removeAddress(ctx, db, ip, input.Serial, input.HwAddress) {
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
