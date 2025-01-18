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

}  // namespace melo::webapi
