package main

import (
	"github.com/hootsuite/atlantis/cmd"
	"github.com/spf13/viper"
)

func main() {
	viper.Set("version", "0.1.0")
	cmd.Execute()
}
