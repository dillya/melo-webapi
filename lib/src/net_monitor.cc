// Copyright (C) 2024 Alexandre Dilly <dillya@sparod.com>

#include "melo/webapi/net_monitor.h"

#include <arpa/inet.h>
#include <asm/types.h>
#include <linux/netlink.h>
#include <linux/rtnetlink.h>
#include <sys/socket.h>

// External dependencies
#include <spdlog/spdlog.h>

namespace melo::webapi {

NetMonitor::NetMonitor(Mode mode, const EventCallback &cb, bool init)
    : mode_(mode), cb_(cb) {
  // Open netlink socket
  ntlk_fd_ = socket(AF_NETLINK, SOCK_RAW, NETLINK_ROUTE);
  if (ntlk_fd_ == -1) {
    SPDLOG_ERROR("failed to open netlink socket: {}", strerror(errno));
    return;
  }

  // Prepare netlink socket options to listen for:
  //  - link changes
  //  - ip changes
  //  - route changes
  struct sockaddr_nl sa {};
  sa.nl_family = AF_NETLINK;
  sa.nl_pid = getpid();
  sa.nl_groups = RTMGRP_LINK | RTMGRP_IPV4_IFADDR;

  // Bind to socket
  if (bind(ntlk_fd_, reinterpret_cast<sockaddr *>(&sa), sizeof(sa))) {
    SPDLOG_ERROR("failed to bind netlink socket: {}", strerror(errno));
  }

  // Initialize list of interfaces
  if (init && !get_link()) {
    SPDLOG_WARN("failed to initialize list of interfaces");
  }

  if (mode_ == Mode::THREAD) {
    // Create the thread
    thread_ = std::jthread(std::bind_front(&NetMonitor::thread_entry, this));
  } else {
    // Set as non-blocking
    int flags = fcntl(ntlk_fd_, F_GETFL, 0);
    if (flags != -1) {
      flags |= O_NONBLOCK;
      flags = fcntl(ntlk_fd_, F_SETFL, flags);
    }
    if (flags == -1) {
      SPDLOG_WARN("failed to set netlink socket as non-blocking");
    }
  }
}

NetMonitor::~NetMonitor() {
  // Stop and join the thread
  if (mode_ == Mode::THREAD) {
    thread_.request_stop();
    thread_.join();
  }

  // Close socket
  if (ntlk_fd_ != -1) {
    close(ntlk_fd_);
  }
}

bool NetMonitor::run_once() {
  struct nlmsghdr buf[8192 / sizeof(nlmsghdr)];

  // Prepare message
  struct sockaddr_nl sa;
  struct iovec iov {
    .iov_base = buf, .iov_len = sizeof(buf),
  };
  struct msghdr msg {
    .msg_name = &sa, .msg_namelen = sizeof(sa), .msg_iov = &iov,
    .msg_iovlen = 1,
  };

  // Get next message from netlink socket
  ssize_t len = recvmsg(ntlk_fd_, &msg, 0);
  if (len <= 0) {
    SPDLOG_ERROR("failed to read the message from netlink socket");
    return false;
  }

  // Process messages one by one
  for (auto *nh = reinterpret_cast<struct nlmsghdr *>(buf); NLMSG_OK(nh, len);
       nh = NLMSG_NEXT(nh, len)) {
    // Process message
    switch (nh->nlmsg_type) {
      case RTM_NEWLINK:
        parse_link(nh, false);
        break;
      case RTM_DELLINK:
        parse_link(nh, true);
        break;
      case RTM_NEWADDR:
        parse_address(nh, false);
        break;
      case RTM_DELADDR:
        parse_address(nh, true);
        break;
      case NLMSG_DONE:
        SPDLOG_DEBUG("[DONE]");
        if (next_msg_ == NextMessage::ADDRESS) {
          get_address();
        }
        break;
      case NLMSG_ERROR:
        SPDLOG_DEBUG("[ERROR]");
        break;
      default:
        break;
    }
  }

  return true;
}

void NetMonitor::parse_link(struct nlmsghdr *nh, bool del) {
  auto *msg = reinterpret_cast<struct ifinfomsg *>(NLMSG_DATA(nh));
  InterfaceInfo info{};

  // Get interface name
  info.index = msg->ifi_index;

  // Extract infos
  struct rtattr *ra = IFLA_RTA(msg);
  int rlen = IFLA_PAYLOAD(nh);
  for (; rlen && RTA_OK(ra, rlen); ra = RTA_NEXT(ra, rlen)) {
    if (ra->rta_type == IFLA_IFNAME) {
      // Get interface name
      info.name = std::string_view(reinterpret_cast<char *>(RTA_DATA(ra)));
    } else if (ra->rta_type == IFLA_ADDRESS) {
      // Get hardware address
      memcpy(info.mac, RTA_DATA(ra), 6);
    }
  }

  SPDLOG_DEBUG("[{} LINK] {} = {}: {}", del ? "DEL" : "NEW", info.index,
               info.name, mac_to_string(info.mac));

  // Call event callback
  if (cb_) {
    cb_(del ? EventType::DEL_INTERFACE : EventType::NEW_INTERFACE, info);
  }
}

void NetMonitor::parse_address(struct nlmsghdr *nh, bool del) {
  auto *msg = reinterpret_cast<struct ifaddrmsg *>(NLMSG_DATA(nh));
  InterfaceInfo info{};

  // Get interface name
  info.index = msg->ifa_index;

  // Extract infos
  struct rtattr *ra = IFLA_RTA(msg);
  int rlen = IFLA_PAYLOAD(nh);
  for (; rlen && RTA_OK(ra, rlen); ra = RTA_NEXT(ra, rlen)) {
    if (ra->rta_type == IFA_LABEL) {
      // Get interface name
      info.name = std::string_view(reinterpret_cast<char *>(RTA_DATA(ra)));
    } else if (ra->rta_type == IFA_LOCAL) {
      // Get IP address
      if (msg->ifa_family == AF_INET) {
        info.ipv4 = *reinterpret_cast<struct in_addr *>(RTA_DATA(ra));
      } else {
        info.ipv6 = *reinterpret_cast<struct in6_addr *>(RTA_DATA(ra));
      }
    }
  }

  SPDLOG_DEBUG("[{} ADDR] {} = {}: {}", del ? "DEL" : "NEW", info.index,
               info.name,
               msg->ifa_family == AF_INET6 ? ip_to_string(info.ipv6)
                                           : ip_to_string(info.ipv4));

  // Call event callback
  if (cb_) {
    cb_(del ? EventType::DEL_ADDRESS : EventType::NEW_ADDRESS, info);
  }
}

bool NetMonitor::get_link() {
  // Prepare RTM_GETLINK request
  struct req {
    struct nlmsghdr hdr;
    struct ifinfomsg pay;
  } req;
  req.hdr.nlmsg_len = NLMSG_LENGTH(sizeof(req.pay));
  req.hdr.nlmsg_type = RTM_GETLINK;
  req.hdr.nlmsg_flags = NLM_F_REQUEST | NLM_F_DUMP;
  req.hdr.nlmsg_pid = getpid();
  req.hdr.nlmsg_seq = seq_.fetch_add(1);
  req.pay.ifi_family = AF_UNSPEC;

  // Socket address
  struct sockaddr_nl sa {};
  sa.nl_family = AF_NETLINK;

  // Create message
  struct iovec iov {
    .iov_base = &req, .iov_len = req.hdr.nlmsg_len,
  };
  struct msghdr msg {
    .msg_name = &sa, .msg_namelen = sizeof(sa), .msg_iov = &iov,
    .msg_iovlen = 1,
  };

  // Send a RTM_GETLINK message
  ssize_t len = sendmsg(ntlk_fd_, &msg, 0);
  if (len <= 0) {
    SPDLOG_ERROR("failed to get current links: {}", strerror(errno));
    return false;
  }

  // After links, get addresses
  next_msg_ = NextMessage::ADDRESS;

  return true;
}

bool NetMonitor::get_address() {
  // Prepare RTM_GETADDR request
  struct req {
    struct nlmsghdr hdr;
    struct ifaddrmsg pay;
  } req;
  req.hdr.nlmsg_len = NLMSG_LENGTH(sizeof(req.pay));
  req.hdr.nlmsg_type = RTM_GETADDR;
  req.hdr.nlmsg_flags = NLM_F_REQUEST | NLM_F_DUMP;
  req.hdr.nlmsg_pid = getpid();
  req.hdr.nlmsg_seq = seq_.fetch_add(1);
  req.pay.ifa_family = AF_INET;

  // Socket address
  struct sockaddr_nl sa {};
  sa.nl_family = AF_NETLINK;

  // Create message
  struct iovec iov {
    .iov_base = &req, .iov_len = req.hdr.nlmsg_len,
  };
  struct msghdr msg {
    .msg_name = &sa, .msg_namelen = sizeof(sa), .msg_iov = &iov,
    .msg_iovlen = 1,
  };

  // Send a RTM_GETADDR message
  ssize_t len = sendmsg(ntlk_fd_, &msg, 0);
  if (len <= 0) {
    SPDLOG_ERROR("failed to get current address: {}", strerror(errno));
    next_msg_ = NextMessage::NONE;
    return false;
  }

  // After addresses, no more messages
  next_msg_ = NextMessage::NONE;

  return true;
}

void NetMonitor::thread_entry(std::stop_token stoken) {
  // Monitor until destruction
  while (!stoken.stop_requested()) {
    // This call is blocking in THREAD mode
    run_once();
  }
}

}  // namespace melo::webapi
