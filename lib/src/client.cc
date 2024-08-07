// Copyright (C) 2025 Alexandre Dilly <dillya@sparod.com>

#include "melo/webapi/client.h"

// External dependencies
#include <curl/curl.h>
#include <spdlog/spdlog.h>

namespace melo::webapi {

Client::Client() {
  // Initialize curl
  auto code = curl_global_init(CURL_GLOBAL_DEFAULT);
  if (code != CURLE_OK) {
    SPDLOG_WARN("failed to initialize curl: {}", curl_easy_strerror(code));
  } else {
    init_ = true;
  }
}

Client::~Client() {
  // Release curl
  if (init_) {
    curl_global_cleanup();
  }
}

}  // namespace melo::webapi
