package main

import (
	"flag"
	"log"
	"os"

	"github.com/jroimartin/gocui"
)

func main() {
	//go pprof.InitPProf()

	flags := initFlag()
	var err error
	if flags.File != "" {
		tree, err = fromFile(flags.File)
	} else {
		if !checkStdInFromPiped() {
			flag.Usage()
			return
		}
		tree, err = fromReader(os.Stdin)
	}
	if err != nil {
		log.Panicln(err)
	}
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	initGUI(g)

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}
