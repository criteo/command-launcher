package frontend

import "strings"

func ParseFlagDefinition(line string) (string, string, string, string, string) {
	flagParts := strings.Split(line, "\t")
	name := strings.TrimSpace(flagParts[0])
	short := ""
	description := ""
	flagType := "string"
	defaultValue := ""
	if len(flagParts) == 2 {
		description = strings.TrimSpace(flagParts[1])
	}
	if len(flagParts) > 2 {
		short = strings.TrimSpace(flagParts[1])
		description = strings.TrimSpace(flagParts[2])
	}
	if len(flagParts) > 3 {
		flagType = strings.TrimSpace(flagParts[3])
	}
	if len(flagParts) > 4 {
		defaultValue = strings.TrimSpace(flagParts[4])
	}

	return name, short, description, flagType, defaultValue
}
