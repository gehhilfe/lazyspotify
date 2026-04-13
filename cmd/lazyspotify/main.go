package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/dubeyKartikay/lazyspotify/buildinfo"
	"github.com/dubeyKartikay/lazyspotify/cli"
	"github.com/dubeyKartikay/lazyspotify/core/utils"
	ui "github.com/dubeyKartikay/lazyspotify/ui/v1"
)

var versionFlag bool

func init() {
	flag.BoolVar(&versionFlag, "version", false, "print build metadata")
}

func main() {
	flag.Parse()
	if versionFlag {
		_ = buildinfo.PrintVersion(os.Stdout)
		return
	}
	if flag.NArg() > 0 && flag.Arg(0) == "version" {
		cli.Run(flag.Args())
		return
	}
	if err := utils.ValidateStartupConfig(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	switch {
	case flag.NArg() > 0:
		cli.Run(flag.Args())
	default:
		ui.RunTui()
	}
}
