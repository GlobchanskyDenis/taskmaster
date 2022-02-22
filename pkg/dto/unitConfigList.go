package dto

type UnitListConfig []*UnitConfig

func (u UnitListConfig) GetMaxStopTime() uint {
	var maxStopTime uint
	for _, conf := range u {
		if conf.Stoptime > maxStopTime {
			maxStopTime = conf.Stoptime
		}
	}
	return maxStopTime
}
