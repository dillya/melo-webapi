// Copyright (C) 2025 Alexandre Dilly <dillya@sparod.com>

#include "melo/webapi/client_helper.h"

#include <fmt/format.h>
#include <nlohmann/json.hpp>

namespace melo::webapi {

static nlohmann::json interface_to_json(const Device::Interface &iface) {
  nlohmann::json obj = {
      {"name", iface.name},
      {"mac", iface.mac},
      {"ipv4", iface.ipv4},
      {"ipv6", iface.ipv6},
  };
  if (iface.type != Device::InterfaceType::kUnknown) {
    obj["type"] = Device::interface_type_to_string(iface.type);
  }
  return obj;
}

void ClientHelper::create_add_device(const Device &dev, bool full,
                                     std::string &method, std::string &url,
                                     std::string &body) {
  method = "PUT";
  url = "/device/add";

  auto &desc = dev.get_description();
  nlohmann::json req = {
      {"serial", desc.serial_number},
      {"name", desc.name},
      {"description", desc.description},
      {"http_port", desc.http_port},
      {"https_port", desc.https_port},
      {"location", desc.location},
      {"online", true},
  };

  if (desc.icon != Device::Icon::kUnknown) {
    req["icon"] = Device::icon_to_string(desc.icon);
  }

  if (full) {
    auto ifaces = nlohmann::json::array();
    for (auto const &iface : dev.get_interface_list()) {
      ifaces.push_back(interface_to_json(iface));
    }
    req["ifaces"] = ifaces;
  }

  body = req.dump();
}

void ClientHelper::create_remove_device(const Device &dev, std::string &method,
                                        std::string &url) {
  method = "DELETE";
  url = fmt::format("/device/{}", dev.get_description().serial_number);
}

void ClientHelper::create_update_device_online_status(const Device &dev,
                                                      bool online,
                                                      std::string &method,
                                                      std::string &url) {
  method = "PUT";
  url = fmt::format("/device/{}/{}", dev.get_description().serial_number,
                    online ? "online" : "offline");
}

void ClientHelper::create_add_device_interface(const Device &dev,
                                               const Device::Interface &iface,
                                               std::string &method,
                                               std::string &url,
                                               std::string &body) {
  method = "PUT";
  url = fmt::format("/device/{}/add", dev.get_description().serial_number);

  auto &desc = dev.get_description();
  body = interface_to_json(iface).dump();
}

void ClientHelper::create_remove_device_interface(const Device &dev,
                                                  const std::string &mac,
                                                  std::string &method,
                                                  std::string &url) {
  method = "DELETE";
  url = fmt::format("/device/{}/{}", dev.get_description().serial_number, mac);
}

bool ClientHelper::generic_parse(int code, const std::string &body,
                                 std::string *error) {
  if (code != 200) {
    if (error) {
      parse_error(body, *error);
    }
    return false;
  }

  return true;
}

void ClientHelper::parse_error(const std::string &body, std::string &error) {
  try {
    auto resp = nlohmann::json::parse(body);
    for (const auto &e : resp["errors"]) {
      error.append(fmt::format("{}: {}", e["location"].get<std::string>(),
                               e["message"].get<std::string>()));
    }
  } catch (const std::exception &e) {
    error = e.what();
  }
}

}  // namespace melo::webapi
