// Copyright (C) 2025 Alexandre Dilly <dillya@sparod.com>

#include "melo/webapi/device.h"

#include <algorithm>

#include <fmt/format.h>
#include <ifaddrs.h>
#include <linux/if_packet.h>
#include <net/if.h>
#include <sys/types.h>

namespace melo::webapi {

std::string Device::get_host_serial_number() {
  struct ifaddrs *ifap;

  // Get network interfaces list
  if (getifaddrs(&ifap)) {
    return {};
  }

  // Get first hardware address for serial */
  std::string serial;
  for (struct ifaddrs *i = ifap; i != nullptr; i = i->ifa_next) {
    // Use only valid interfaces
    if (i->ifa_addr->sa_family != AF_PACKET || (i->ifa_flags & IFF_LOOPBACK)) {
      continue;
    }

    // Extract MAC address
    // TODO(dillya): the order is not guaranteed?
    struct sockaddr_ll *s = (struct sockaddr_ll *)i->ifa_addr;
    auto &mac = s->sll_addr;
    if (std::all_of(std::begin(mac), std::end(mac),
                    [](const auto &val) { return val == 0; })) {
      continue;
    }

    // Generate serial number
    serial = fmt::format("{:02x}:{:02x}:{:02x}:{:02x}:{:02x}:{:02x}", mac[0],
                         mac[1], mac[2], mac[3], mac[4], mac[5]);
    break;
  }

  // Free interfaces list
  freeifaddrs(ifap);

  return serial;
}

Device::Interface *Device::add_interface(Interface iface) {
  if (iface.mac.empty()) {
    return nullptr;
  }

  // Update current interface
  auto *current = get_interface(iface.mac);
  if (current) {
    *current = std::move(iface);
    return current;
  }

  // Add interface
  return &ifaces_.emplace_back(std::move(iface));
}

bool Device::remove_interface(const std::string &mac) {
  if (mac.empty()) {
    return false;
  }

  // Find interface
  auto *current = get_interface(mac);
  if (!current) {
    return false;
  }

  // Remove interface
  auto pos = current - &(*ifaces_.begin());
  ifaces_.erase(ifaces_.begin() + pos);

  return false;
}

std::string_view Device::icon_to_string(Icon icon) {
  switch (icon) {
    case Icon::kLivingRoom:
      return "living";
    case Icon::kKitchen:
      return "kitchen";
    case Icon::kBedroom:
      return "bed";
    case Icon::kUnknown:
    default:
      return "unknown";
  }
}

std::string_view Device::interface_type_to_string(InterfaceType type) {
  switch (type) {
    case InterfaceType::kEthernet:
      return "ethernet";
    case InterfaceType::kWifi:
      return "wifi";
    case InterfaceType::kUnknown:
    default:
      return "unknown";
  }
}

}  // namespace melo::webapi
