package color

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/charmbracelet/lipgloss"
)

var containerNameColors = []lipgloss.Color{
	lipgloss.Color("#58A2EE"), // blue
	lipgloss.Color("#3FE34B"), // bright green
	lipgloss.Color("#7c60d7"), // purple
	lipgloss.Color("#FD2C4C"), // red
	lipgloss.Color("#FE7A00"), // orange
	lipgloss.Color("#FAF81C"), // yellow
	lipgloss.Color("#56EBD3"), // teal
	lipgloss.Color("#42952E"), // green
	lipgloss.Color("#FFACE6"), // light pink
	lipgloss.Color("#FE16F4"), // bright pink
	lipgloss.Color("#D6A112"), // gold
	lipgloss.Color("#FFDAB9"), // beige
	lipgloss.Color("#FF7E6A"), // tomato
}

func ContainerColor(name string) lipgloss.Color {
	hash := md5.Sum([]byte(name))
	hashStr := hex.EncodeToString(hash[:])
	var hashValue int64
	_, err := fmt.Sscanf(hashStr[:8], "%x", &hashValue)
	if err != nil {
		return containerNameColors[0]
	}
	return containerNameColors[hashValue%int64(len(containerNameColors))]
}
