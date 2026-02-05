package utils

func TimeToMinutes(hour, minute int) int {
	return hour*60 + minute
}

func MinutesToTime(m int) (hour, minute int) {
	hour = m / 60
	minute = m % 60
	return
}
