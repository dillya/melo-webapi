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
 * @file client_helper.h
 * @brief Web API client heper.
 */

#ifndef MELO_WEBAPI_CLIENT_HELPER_H_
#define MELO_WEBAPI_CLIENT_HELPER_H_

#include <vector>

#include <melo/webapi/device.h>

namespace melo::webapi {

class ClientHelper {
 public:
  /**
   * Create request to add / update a device.
   *
   * @param[in] dev The device
   * @param[in] full Set to `true` to send interface list as well, `false`
   * otherwise
   * @param[out] method The HTTP method to use
   * @param[out] url The HTTP URL to use
   * @param[out] body The HTTP request body to use
   */
  static void create_add_device(const Device &dev, bool full,
                                std::string &method, std::string &url,
                                std::string &body);

  /**
   * Parse response from adding / updating a device.
   *
   * @param[in] code The HTTP response code
   * @param[in] body The HTTP response body to parse
   * @param[out] error The optional returned by the parser
   * @return `true` if the device has been added / updated, `false` otherswise.
   */
  static inline bool parse_add_device(int code, const std::string &body,
                                      std::string *error = nullptr) {
    return generic_parse(code, body, error);
  }

  /**
   * Create request to remove a device.
   *
   * @param[in] dev The device
   * @param[out] method The HTTP method to use
   * @param[out] url The HTTP URL to use
   */
  static void create_remove_device(const Device &dev, std::string &method,
                                   std::string &url);

  /**
   * Parse response from removing a device.
   *
   * @param[in] code The HTTP response code
   * @param[in] body The HTTP response body to parse
   * @param[out] error The optional returned by the parser
   * @return `true` if the device has been removed, `false` otherswise.
   */
  static inline bool parse_remove_device(int code, const std::string &body,
                                         std::string *error = nullptr) {
    return generic_parse(code, body, error);
  }

  /**
   * Create request to update online status of a device.
   *
   * @param[in] dev The device
   * @param[in] online Set to `true` if the device is online, `false` otherwise
   * @param[out] method The HTTP method to use
   * @param[out] url The HTTP URL to use
   */
  static void create_update_device_online_status(const Device &dev, bool online,
                                                 std::string &method,
                                                 std::string &url);

  /**
   * Parse response from updating online status of a device.
   *
   * @param[in] code The HTTP response code
   * @param[in] body The HTTP response body to parse
   * @param[out] error The optional returned by the parser
   * @return `true` if the device has been updated, `false` otherswise.
   */
  static inline bool parse_update_device_online_status(
      int code, const std::string &body, std::string *error = nullptr) {
    return generic_parse(code, body, error);
  }

  /**
   * Create request to add / update a device interface.
   *
   * @param[in] dev The device
   * @param[in] iface The device interface
   * @param[out] method The HTTP method to use
   * @param[out] url The HTTP URL to use
   * @param[out] body The HTTP request body to use
   */
  static void create_add_device_interface(const Device &dev,
                                          const Device::Interface &iface,
                                          std::string &method, std::string &url,
                                          std::string &body);

  /**
   * Parse response from adding / updating a device interface.
   *
   * @param[in] code The HTTP response code
   * @param[in] body The HTTP response body to parse
   * @param[out] error The optional returned by the parser
   * @return `true` if the interface has been added / updated, `false`
   * otherswise.
   */
  static inline bool parse_add_device_interface(int code,
                                                const std::string &body,
                                                std::string *error = nullptr) {
    return generic_parse(code, body, error);
  }

  /**
   * Create request to remove a device interface.
   *
   * @param[in] dev The device
   * @param[in] iface The device interface
   * @param[out] method The HTTP method to use
   * @param[out] url The HTTP URL to use
   */
  static inline void create_remove_device_interface(
      const Device &dev, const Device::Interface &iface, std::string &method,
      std::string &url) {
    create_remove_device_interface(dev, iface.mac, method, url);
  }

  /**
   * Create request to remove a device interface.
   *
   * @param[in] dev The device
   * @param[in] mac The device interface MAC address
   * @param[out] method The HTTP method to use
   * @param[out] url The HTTP URL to use
   */
  static void create_remove_device_interface(const Device &dev,
                                             const std::string &mac,
                                             std::string &method,
                                             std::string &url);

  /**
   * Parse response from removing a device interface.
   *
   * @param[in] code The HTTP response code
   * @param[in] body The HTTP response body to parse
   * @param[out] error The optional returned by the parser
   * @return `true` if the interface has been removed, `false` otherswise.
   */
  static inline bool parse_remove_device_interface(
      int code, const std::string &body, std::string *error = nullptr) {
    return generic_parse(code, body, error);
  }

 private:
  static bool generic_parse(int code, const std::string &body,
                            std::string *error);

  static void parse_error(const std::string &body, std::string &error);
};

}  // namespace melo::webapi

#endif  // MELO_WEBAPI_CLIENT_HELPER_H_
