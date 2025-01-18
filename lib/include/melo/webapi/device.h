// Copyright (C) 2025 Alexandre Dilly <dillya@sparod.com>
//
// This program is free software: you can redistribute it and/or modify it under
// the terms of the GNU Affero General Public License as published by the Free
// Software Foundation, either version 3 of the License, or (at your option) any
// later version.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS
// FOR A PARTICULAR PURPOSE. See the GNU Affero General Public License for more
// details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

/**
 * @file device.h
 * @brief Device definition.
 */

#ifndef MELO_WEBAPI_DEVICE_H_
#define MELO_WEBAPI_DEVICE_H_

#include <string>
#include <vector>

namespace melo::webapi {

/**
 * Melo device class.
 */
class Device {
 public:
  /**
   * Device Icon.
   */
  enum class Icon {
    kUnknown,     //!< Unknown icon
    kLivingRoom,  //!< Living room
    kKitchen,     //!< Kitchen
    kBedroom,     //!< Bedroom
  };

  /**
   * Interface type.
   */
  enum class InterfaceType {
    kUnknown,   //!< Unknown interface
    kEthernet,  //!< Ethernet interface
    kWifi,      //!< WiFi interface
  };

  /**
   * Description of a device.
   */
  struct Descriptor {
    std::string serial_number;  //!< The serial number of the device
    std::string name;           //!< The name of the device
    std::string description;    //!< The description of the device
    Icon icon;                  //!< The icon used to represent the device
    std::string location;       //!< Location of the device

    uint16_t http_port;   //!< HTTP port of the device
    uint16_t https_port;  //!< HTTPs port of the device (0: disabled)

    /**
     * Create a default descriptor.
     *
     * @param[in] serial The serial number of the device
     * @param[in] name The name of the device
     * @param[in] http_port The HTTP port of the device
     * @param[in] https_port The HTTPs port of the device (0: disabled)
     */
    Descriptor(std::string serial, std::string name = "Melo",
               uint16_t http_port = 8080, uint16_t https_port = 0)
        : serial_number(std::move(serial)),
          name(std::move(name)),
          description(),
          icon(Icon::kUnknown),
          location(""),
          http_port(http_port),
          https_port(https_port) {}
  };

  /**
   * Description of an interface.
   */
  struct Interface {
    InterfaceType type;  //!< Type of the interface
    std::string name;    //!< Name of the interface
    std::string mac;     //!< MAC address of the interface
    std::string ipv4;    //!< IPv4 address of the interface
    std::string ipv6;    //!< IPv6 address of the interface
  };

  /**
   * Create a new device.
   */
  Device() = default;

  /**
   * Create a new device.
   *
   * @param[in] desc The description of the device
   */
  Device(Descriptor desc) : desc_(std::move(desc)) {}

  /**
   * Get description of the device.
   *
   * @return the description of the device.
   */
  [[nodiscard]] inline const Descriptor &get_description() const {
    return desc_;
  }

  /**
   * Update description of the device.
   *
   * @param[in] desc The updated description of the device
   */
  inline void update_description(Descriptor desc) { desc_ = std::move(desc); }

  /**
   * Get interface list.
   *
   * @return the interface list.
   */
  [[nodiscard]] inline const std::vector<Interface> &get_interface_list()
      const {
    return ifaces_;
  }

  /**
   * Get interface by MAC address.
   *
   * @param[in] mac The MAC address of the interface
   * @return a reference to the interface or `nullptr` if not existing.
   */
  [[nodiscard]] inline Interface *get_interface(const std::string &mac) {
    for (auto &iface : ifaces_) {
      if (iface.mac == mac) {
        return &iface;
      }
    }
    return nullptr;
  }

  /**
   * Get interface by name.
   *
   * @param[in] name The name of the interface
   * @return a reference to the interface or `nullptr` if not existing.
   */
  [[nodiscard]] inline Interface *get_interface_by_name(
      const std::string &name) {
    for (auto &iface : ifaces_) {
      if (iface.name == name) {
        return &iface;
      }
    }
    return nullptr;
  }

  /**
   * Add / update an interface.
   *
   * The interface is added to the list if it doesn't exist yet or the full
   *structure will be updated.
   *
   * @return a pointer to the interface structure internally stored (valid until
   *remove_interface() is called), `nullptr` otherwise.
   */
  Interface *add_interface(Interface iface);

  /**
   * Remove an interface.
   *
   * @return `true` if the interface has been removed, `false` otherwise.
   */
  bool remove_interface(const std::string &mac);

  /**
   * Get host serial number.
   *
   * This function will provide the unique serial number of the current machine.
   *
   * @return a unique serial number identifying the current machine.
   */
  static std::string get_host_serial_number();

  /**
   * Convert Icon to string.
   *
   * @param[in] icon The icon to convert
   * @return the string corresponding to the icon or "unknown".
   */
  static std::string_view icon_to_string(Icon icon);

  /**
   * Convert Interface type to string.
   *
   * @param[in] type The interface type to convert
   * @return the string corresponding to the interface type or "unknown".
   */
  static std::string_view interface_type_to_string(InterfaceType type);

 private:
  Descriptor desc_;                //!< Description of the device
  std::vector<Interface> ifaces_;  //!< Interfaces of the device
};

}  // namespace melo::webapi

#endif  // MELO_WEBAPI_DEVICE_H_
