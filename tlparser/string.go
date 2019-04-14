package tlparser

import "strings"

// FixAAA fixes known acronyms and abbreviations to be upper case.
func FixAAA(str string) string {
	result := str
	result = strings.ReplaceAll(result, "Api", "API")
	result = strings.ReplaceAll(result, "Id", "ID")
	result = strings.ReplaceAll(result, "IDenti", "Identi")
	result = strings.ReplaceAll(result, "Ip", "IP")
	result = strings.ReplaceAll(result, "Json", "JSON")
	result = strings.ReplaceAll(result, "Html", "HTML")
	result = strings.ReplaceAll(result, "Http", "HTTP")
	result = strings.ReplaceAll(result, "Ttl", "TTL")
	result = strings.ReplaceAll(result, "Udp", "UDP")
	result = strings.ReplaceAll(result, "Uri", "URI")
	result = strings.ReplaceAll(result, "Url", "URL")
	return result
}
