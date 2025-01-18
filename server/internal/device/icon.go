package device

type Icon uint

const (
  UnknownIcon Icon = iota
  LivingIcon
  KitchenIcon
  BedIcon
)

var iconMap = [...]string{"unknown", "living","kitchen","bed"}

func (i Icon) ToString() string {
	if int(i) < len(iconMap) {
		return iconMap[i]
	}
	return iconMap[0]
}

func (i Icon) FromString(str string) uint {
	for index := range iconMap {
		if iconMap[index] == str {
			return uint(index)
		}
	}
	return 0
}
