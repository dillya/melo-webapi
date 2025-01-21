import json
import unittest
from urllib.request import urlopen
from typing import Any, Dict

class TestLegcayAPI(unittest.TestCase):
    _url_base = "http://127.0.0.1:8888/discover"

    def _do_request(self, action: str, extra: Dict[str, str] = {}, status: int = 200) -> Any:
        url = f"{self._url_base}?action={action}"
        for key, value in extra.items():
            url += f"&{key}={value}"
        with urlopen(url) as response:
            self.assertEqual(response.status, status)
            return json.loads(response.read())

    def test_device_managment(self) -> None:
        serial = "01:23:45:67:89:ab"
        name = "test"
        port = 1234

        # Add device
        resp = self._do_request("add_device", {"serial": serial, "name": name, "port": str(port)})
        self.assertEqual(resp, {})

        # List device
        resp = self._do_request("list")
        self.assertEqual(len(resp), 1)
        self.assertEqual(resp[0]["serial"], serial)
        self.assertEqual(resp[0]["name"], name)
        self.assertEqual(resp[0]["port"], port)

        # Update device
        name = "new_test"
        port = 8080
        resp = self._do_request("add_device", {"serial": serial, "name": name, "port": str(port)})
        self.assertEqual(resp, {})

        # List device
        resp = self._do_request("list")
        self.assertEqual(len(resp), 1)
        self.assertEqual(resp[0]["name"], name)
        self.assertEqual(resp[0]["port"], port)

        # Remove device
        resp = self._do_request("remove_device", {"serial": serial})
        self.assertEqual(resp, {})

        # List device
        resp = self._do_request("list")
        self.assertEqual(len(resp), 0)

    def test_interface_managment(self) -> None:
        serial = "01:23:45:67:89:ab"
        name = "test"
        port = 1234

        # Add device
        resp = self._do_request("add_device", {"serial": serial, "name": name, "port": str(port)})
        self.assertEqual(resp, {})

        hw_address = "cd:ef:98:76:54:32"
        address="192.168.0.20"

        # Add interface
        resp = self._do_request("add_address", {"serial": serial, "hw_address": hw_address, "address": address})
        self.assertEqual(resp, {})

        # List device
        resp = self._do_request("list")
        self.assertEqual(len(resp), 1)
        self.assertEqual(resp[0]["serial"], serial)
        ifaces = resp[0]["list"]
        self.assertEqual(len(ifaces), 1)
        self.assertEqual(ifaces[0]["hw_address"], hw_address)
        self.assertEqual(ifaces[0]["address"], address)

        # Update address
        address="192.168.0.10"
        resp = self._do_request("add_address", {"serial": serial, "hw_address": hw_address, "address": address})
        self.assertEqual(resp, {})

        # List device
        resp = self._do_request("list")
        self.assertEqual(len(resp), 1)
        self.assertEqual(resp[0]["list"][0]["address"], address)

        # Update address
        resp = self._do_request("remove_address", {"serial": serial, "hw_address": hw_address})
        self.assertEqual(resp, {})

        # List device
        resp = self._do_request("list")
        self.assertEqual(len(resp), 1)
        self.assertEqual(len(resp[0]["list"]), 0)

        # Remove device
        resp = self._do_request("remove_device", {"serial": serial})
        self.assertEqual(resp, {})

        # List device
        resp = self._do_request("list")
        self.assertEqual(len(resp), 0)

if __name__ == '__main__':
    unittest.main()
