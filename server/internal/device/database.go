package device

import (
	"context"
	"database/sql"
	"time"

	"github.com/dillya/melo-webapi/internal/utils"

	log "github.com/sirupsen/logrus"
)

func InitializeTables(db *sql.DB) bool {
	const version = 1

	// Get version
	table_version := utils.GetTableVersion(db, "device")
	if table_version == version {
		return true
	}

	log.Infof("recreate Device tables due to update: %d -> %d", table_version, version)

	// Remove previous tables
	_, err := db.Exec("DROP TABLE IF EXISTS device_iface, device CASCADE")
	if err != nil {
		log.Errorf("failed to drop old tables: %s", err)
		return false
	}

	// Create device table
	device := `CREATE TABLE device (
  id INT(11) NOT NULL AUTO_INCREMENT,
  ip INT(10) unsigned NOT NULL,
  serial VARCHAR(17) NOT NULL,
  name VARCHAR(128) NOT NULL,
  description VARCHAR(256),
  icon TINYINT(3) unsigned NOT NULL DEFAULT 0,
  location VARCHAR(128),
  http_port MEDIUMINT(9) NOT NULL,
  https_port MEDIUMINT(9) NOT NULL DEFAULT 0,
  online BOOL DEFAULT FALSE,
  last_update BIGINT(4) UNSIGNED NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY serial_ip (serial,ip),
  KEY serial (serial),
  KEY ip (ip)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_uca1400_ai_ci;`
	_, err = db.Exec(device)
	if err != nil {
		log.Errorf("failed to create device table: %s", err)
		return false
	}

	// Create device_iface table
	device_iface := `CREATE TABLE device_iface (
  id INT(11) NOT NULL AUTO_INCREMENT,
  device_id INT(11) NOT NULL,
  ipv4 INT(10) UNSIGNED,
  ipv6 VARBINARY(16),
  mac BIGINT(20) UNSIGNED NOT NULL,
  name VARCHAR(128) NOT NULL DEFAULT 'Unknown',
  type INT(11) NOT NULL DEFAULT 0,
  PRIMARY KEY (id),
  UNIQUE KEY device_id_mac (device_id,mac),
  CONSTRAINT device_iface_constraint FOREIGN KEY (device_id) REFERENCES device (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_uca1400_ai_ci;`
	_, err = db.Exec(device_iface)
	if err != nil {
		log.Errorf("failed to create device interface table: %s", err)
		return false
	}

	// Update version
	return utils.UpdateTableVersion(db, "device", 1)
}

func listInterface(ctx context.Context, db *sql.DB, id uint) []DeviceInterface {
	// Create interface list
	list := []DeviceInterface{}

	// Fetch interfaces of the current device
	ifaces, err := db.QueryContext(ctx, "SELECT type, name, mac, INET_NTOA(ipv4), INET6_NTOA(ipv6) FROM device_iface WHERE device_id=?", id)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("failed to get interface list")
		return list
	}
	defer ifaces.Close()

	// Generate list
	for ifaces.Next() {
		// Scan interface
		var iface_type uint
		var mac uint64
		var name string
		var ipv4, ipv6 []byte
		if err := ifaces.Scan(&iface_type, &name, &mac, &ipv4, &ipv6); err != nil {
			log.WithFields(log.Fields{"error": err}).Error("failed to scan interface")
			continue
		}

		// Add interface to list
		list = append(list, DeviceInterface{
			Type:        InterfaceType.ToString(InterfaceType(iface_type)),
			Name:        name,
			MacAddress:  utils.Uint64ToHwAddress(mac),
			Ipv4Address: string(ipv4),
			Ipv6Address: string(ipv6),
		})
	}

	return list
}

func List(ctx context.Context, db *sql.DB, ip string) []Device {
	// Create device list
	list := []Device{}

	// Fetch devices
	devices, err := db.QueryContext(ctx, "SELECT id, name, serial, description, icon, location, http_port, https_port, online, last_update FROM device WHERE ip=INET_ATON(?)", ip)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("failed to get device list")
		return list
	}
	defer devices.Close()

	// Generate list
	for devices.Next() {
		// Scan device
		var online bool
		var id, icon uint
		var http_port, https_port uint16
		var last_update uint64
		var serial, name string
		var description, location []byte
		if err := devices.Scan(&id, &name, &serial, &description, &icon, &location, &http_port, &https_port, &online, &last_update); err != nil {
			log.WithFields(log.Fields{"error": err}).Error("failed to scan device")
			continue
		}

		// Add device to list
		list = append(list, Device{
			DeviceDesc: DeviceDesc{
				Serial:      serial,
				Name:        name,
				Description: string(description),
				Icon:        Icon.ToString(Icon(icon)),
				Location:    string(location),
				HttpPort:    http_port,
				HttpsPort:   https_port,
				Online:      online,
			},
			LastUpdate: last_update,
			Interfaces: listInterface(ctx, db, id),
		})
	}

	return list
}

func Add(ctx context.Context, db *sql.DB, ip string, desc DeviceDesc) bool {
	// Add or update device
	ts := time.Now().Unix()
	result, err := db.ExecContext(ctx, `INSERT INTO device
(ip, serial, name, description, icon, location, http_port, https_port, online, last_update)
VALUES (INET_ATON(?), ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE name=?, description=?, icon=?, location=?, http_port=?, https_port=?, online=?, last_update=?`,
		ip,
		desc.Serial,
		desc.Name,
		desc.Description,
		IconFromString(desc.Icon),
		desc.Location,
		desc.HttpPort,
		desc.HttpsPort,
		desc.Online,
		ts,
		desc.Name,
		desc.Description,
		IconFromString(desc.Icon),
		desc.Location,
		desc.HttpPort,
		desc.HttpsPort,
		desc.Online,
		ts,
	)
	if err != nil {
		log.WithFields(log.Fields{"error": err, "device": desc}).Error("failed to add device")
		return false
	}
	_, err = result.RowsAffected()
	return err == nil
}

func Remove(ctx context.Context, db *sql.DB, ip string, serial string) bool {
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

func UpdateStatus(ctx context.Context, db *sql.DB, ip string, serial string, online bool) bool {
	// Update status
	ts := time.Now().Unix()
	_, err := db.Exec("UPDATE device SET online=?, last_update = ? WHERE ip = INET_ATON(?) AND serial=?", online, ts, ip, serial)
	if err != nil {
		log.WithFields(log.Fields{"error": err, "serial": serial}).Error("failed to update device status")
	}
	return err == nil
}

func AddAddress(ctx context.Context, db *sql.DB, ip string, serial string, iface DeviceInterface) bool {
	// Add or update address
	result, err := db.ExecContext(ctx, `INSERT INTO device_iface
(device_id, mac, type, name, ipv4, ipv6)
SELECT id, ?, ?, ?, INET_ATON(?), INET6_ATON(?)
FROM device WHERE ip=INET_ATON(?) AND serial=?
ON DUPLICATE KEY UPDATE type=?, name=?, ipv4=INET_ATON(?), ipv6=INET6_ATON(?)`,
		utils.Uint64FromHwAddress(iface.MacAddress),
		InterfaceTypeFromString(iface.Type),
		iface.Name,
		iface.Ipv4Address,
		iface.Ipv6Address,
		ip,
		serial,
		InterfaceTypeFromString(iface.Type),
		iface.Name,
		iface.Ipv4Address,
		iface.Ipv6Address,
	)
	if err != nil {
		log.WithFields(log.Fields{"error": err, "serial": serial}).Error("failed to add address")
		return false
	}
	_, err = result.RowsAffected()

	// Update device
	UpdateStatus(ctx, db, ip, serial, true)

	return err == nil
}

func RemoveAddress(ctx context.Context, db *sql.DB, ip string, serial string, hw_address string) bool {
	// Remove address
	result, err := db.ExecContext(ctx, "DELETE FROM device_iface WHERE device_id IN (SELECT id FROM device WHERE ip=INET_ATON(?) AND serial=?) AND mac=?\n",
		ip,
		serial,
		utils.Uint64FromHwAddress(hw_address),
	)
	if err != nil {
		log.WithFields(log.Fields{"error": err, "serial": serial}).Error("failed to remove address")
		return false
	}
	rows, err := result.RowsAffected()

	// Update device
	UpdateStatus(ctx, db, ip, serial, true)

	return err == nil && rows == 1
}
