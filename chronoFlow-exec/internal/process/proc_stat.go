package process

import (
	"fmt"
	"strconv"
	"strings"
)

func parseProcStatProcessGroupID(stat string) (int, error) {
	end := strings.LastIndex(stat, ")")
	if end < 0 {
		return 0, fmt.Errorf("invalid proc stat: missing process name")
	}
	fields := strings.Fields(stat[end+1:])
	if len(fields) < 3 {
		return 0, fmt.Errorf("invalid proc stat: missing process group")
	}
	pgid, err := strconv.Atoi(fields[2])
	if err != nil {
		return 0, fmt.Errorf("invalid proc stat process group: %w", err)
	}
	return pgid, nil
}
