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
 * @file client.h
 * @brief Web API client.
 */

#ifndef MELO_WEBAPI_CLIENT_H_
#define MELO_WEBAPI_CLIENT_H_

#include <melo/webapi/device.h>

namespace melo::webapi {

/**
 * Web API client class
 *
 * This class can be used to communicate with the Melo Web API server.
 */
class Client {
 public:
  /**
   * Create a new client.
   *
   * This call will initialize all internal resources such as the HTTP client.
   */
  Client();

  /**
   * Destructor for client.
   *
   * This function will release all internal resources.
   */
  ~Client();

 private:
  bool init_ = false;  //!< Internal resources are initialized
};

}  // namespace melo::webapi

#endif  // MELO_WEBAPI_CLIENT_H_
