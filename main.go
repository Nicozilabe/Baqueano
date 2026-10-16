package main

import (
	"github.com/Nicozilabe/baqueano/cmd"
	"github.com/Nicozilabe/baqueano/internal/updater"
)

func main() {
	updater.CleanupOldBinary()
	cmd.Execute()
}
