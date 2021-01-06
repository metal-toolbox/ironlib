package utils

import (
	"fmt"
	"strings"

	"github.com/packethost/ironlib/model"
)

func vendorFromString(s string) string {

	switch {
	case strings.Contains(s, "LSI3008-IT"):
		return "LSI"
	case strings.Contains(s, "HGST "):
		return "HGST"
	case strings.Contains(s, "Micron_"):
		return "Micron"
	case strings.Contains(s, "TOSHIBA"):
		return "Toshiba"
	case strings.Contains(s, "ConnectX4LX"):
		return "Mellanox"
	default:
		return "unknown"
	}
}

// return the given string with the idx prefixed
func prefixIndex(idx int, s string) string {
	return fmt.Sprintf("[%d] %s", idx, s)
}

func purgeTestComponentID(components []*model.Component) []*model.Component {
	for _, c := range components {
		c.ID = ""
	}
	return components
}
