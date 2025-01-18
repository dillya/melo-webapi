package device

type InterfaceType uint

const (
  UnknownInterface InterfaceType = iota
  EthernetInterface
  WifiInterface
)

var ifaceTypeMap = [...]string{"unknown", "ethernet","wifi"}

func (i InterfaceType) ToString() string {
	if int(i) < len(ifaceTypeMap) {
		return ifaceTypeMap[i]
	}
	return ifaceTypeMap[0]
}

func (i InterfaceType) FromString(str string) uint {
	for index := range ifaceTypeMap {
		if ifaceTypeMap[index] == str {
			return uint(index)
		}
	}
	return 0
}
