package utils

func CheckNA(str string) string {
	if str == "" {
		return "N/A"
	}
	return str
}
