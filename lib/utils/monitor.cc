// Copyright (C) 2024 Alexandre Dilly <dillya@sparod.com>

#include <iostream>

#include <poll.h>
#include <spdlog/cfg/env.h>

#include <melo/webapi/net_monitor.h>

using namespace melo::webapi;

int main(int argc, char *argv[]) {
  // Setup logs
  spdlog::cfg::load_env_levels();

  // Create event callback
  auto cb = [](NetMonitor::EventType type,
               const NetMonitor::InterfaceInfo &info) {
    std::cout << " -> " << info.name << " (" << info.index << "): ";
    switch (type) {
      case NetMonitor::EventType::NEW_INTERFACE:
        std::cout << "MAC = " << NetMonitor::mac_to_string(info.mac)
                  << std::endl;
        break;
      case NetMonitor::EventType::DEL_INTERFACE:
        std::cout << "DELETED" << std::endl;
        break;
      case NetMonitor::EventType::NEW_ADDRESS:
        std::cout << "IP = " << NetMonitor::ip_to_string(info.ipv4)
                  << std::endl;
        break;
      case NetMonitor::EventType::DEL_ADDRESS:
        std::cout << "DISCONNECTED" << std::endl;
        break;
      default:
        std::cout << "unsupported event" << std::endl;
    }
  };

  // Create a new monitor
  NetMonitor monitor{NetMonitor::Mode::POLL, cb, true};

  std::cout << "Start monitoring..." << std::endl;

  pollfd fds{
      .fd = monitor.get_fd(),
      .events = POLLIN,
  };

  // Loop forever
  while (true) {
    // Wait for new events
    int ret = poll(&fds, 1, -1);
    if (ret <= 0) {
      break;
    }

    // Handle events
    if (fds.revents && !monitor.run_once()) {
      std::cerr << "An error occured while monitoring: exit" << std::endl;
      break;
    }
  }

  return 0;
}
