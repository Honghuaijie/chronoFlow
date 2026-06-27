package process

import (
	"fmt"
	"strconv"
	"strings"
)

type procStatInfo struct {
	State          string
	ProcessGroupID int
	Zombie         bool
}

func parseProcStatProcessGroupID(stat string) (int, error) {
	info, err := parseProcStat(stat)
	if err != nil {
		return 0, err
	}
	return info.ProcessGroupID, nil
}

func parseProcStat(stat string) (procStatInfo, error) {
	end := strings.LastIndex(stat, ")")
	if end < 0 {
		return procStatInfo{}, fmt.Errorf("invalid proc stat: missing process name")
	}
	fields := strings.Fields(stat[end+1:])
	if len(fields) < 3 {
		return procStatInfo{}, fmt.Errorf("invalid proc stat: missing process group")
	}
	pgid, err := strconv.Atoi(fields[2])
	if err != nil {
		return procStatInfo{}, fmt.Errorf("invalid proc stat process group: %w", err)
	}
	state := fields[0]
	return procStatInfo{State: state, ProcessGroupID: pgid, Zombie: state == "Z"}, nil
}
