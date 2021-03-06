package cmd

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/dominicbreuker/pspy/internal/config"
	"github.com/dominicbreuker/pspy/internal/fswatcher"
	"github.com/dominicbreuker/pspy/internal/logging"
	"github.com/dominicbreuker/pspy/internal/pspy"
	"github.com/dominicbreuker/pspy/internal/psscanner"
	"github.com/spf13/cobra"
)

var bannerLines = []string{
	"      _____   _____ _______     __",
	"     |  __ \\ / ____|  __ \\ \\   / /",
	"     | |__) | (___ | |__) \\ \\_/ / ",
	"     |  ___/ \\___ \\|  ___/ \\   /  ",
	"     | |     ____) | |      | |   ",
	"     |_|    |_____/|_|      |_|   ",
	helpText,
}

var helpText = `
pspy monitors the system for file system events and new processes.
It prints out these envents to the console.
File system events are monitored with inotify.
Processes are monitored by scanning /proc, using file system events as triggers.
pspy does not require root permissions do operate.
Check our https://github.com/dominicbreuker/pspy for more information.
`

var banner = strings.Join(bannerLines, "\n")

var rootCmd = &cobra.Command{
	Use:   "pspy",
	Short: "pspy can watch your system for new processes and file system events",
	Long:  banner,
	Run:   root,
}

var logPS, logFS bool
var rDirs, dirs []string
var defaultRDirs = []string{
	"/usr",
	"/tmp",
	"/etc",
	"/home",
	"/var",
	"/opt",
}
var defaultDirs = []string{}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&logPS, "procevents", "p", true, "print new processes to stdout")
	rootCmd.PersistentFlags().BoolVarP(&logFS, "fsevents", "f", false, "print file system events to stdout")
	rootCmd.PersistentFlags().StringArrayVarP(&rDirs, "recursive_dirs", "r", defaultRDirs, "watch these dirs recursively")
	rootCmd.PersistentFlags().StringArrayVarP(&dirs, "dirs", "d", defaultDirs, "watch these dirs")

	log.SetOutput(os.Stdout)
}

func root(cmd *cobra.Command, args []string) {
	logger := logging.NewLogger()

	cfg := &config.Config{
		RDirs:        rDirs,
		Dirs:         dirs,
		LogPS:        logPS,
		LogFS:        logFS,
		DrainFor:     1 * time.Second,
		TriggerEvery: 100 * time.Millisecond,
	}
	fsw := fswatcher.NewFSWatcher()
	defer fsw.Close()

	pss := psscanner.NewPSScanner()

	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	b := &pspy.Bindings{
		Logger: logger,
		FSW:    fsw,
		PSS:    pss,
	}
	exit := pspy.Start(cfg, b, sigCh)
	<-exit
	os.Exit(0)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
