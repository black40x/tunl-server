package tui

import (
	"fmt"
)

const (
	Reset       = "\033[0m"
	Red         = "\033[31m"
	Green       = "\033[32m"
	Yellow      = "\033[33m"
	Blue        = "\033[34m"
	Magenta     = "\033[35m"
	Cyan        = "\033[36m"
	White       = "\033[37m"
	BlueBold    = "\033[34;1m"
	MagentaBold = "\033[35;1m"
	RedBold     = "\033[31;1m"
	YellowBold  = "\033[33;1m"
)

func PrintServerStarted(host, port, ver string) {
	fmt.Printf("🚀 "+BlueBold+"Server started at %s:%s"+Reset+"\n", host, port)
	fmt.Printf("Version %s\n", ver)
	fmt.Printf("Visit %s for more information\n\n", "https://github.com/black40x/tunl-server")
}

func PrintError(err error) {
	fmt.Printf(RedBold+"Error: "+Reset+"%s\n", err.Error())
}

func PrintInfo(s string) {
	fmt.Printf(BlueBold+"Info: "+Reset+"%s\n", s)
}

func PrintWarning(s string) {
	fmt.Printf(YellowBold+"Warning: "+Reset+"%s\n", s)
}

func PrintLn(s string) {
	fmt.Println(s)
}
