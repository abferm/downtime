package downtime

import (
	"fmt"
	"os"
	"runtime/debug"
	"time"

	"github.com/lestrrat-go/strftime"
)

// StrftimeToGo converts a c-style strftime format string to the format expected by time.Time.Format()
func StrftimeToGo(cFormat string) (goFormat string, err error) {
	refTime := time.Date(2006, time.January, 2, 15, 4, 5, 999999999, time.FixedZone("MST", -7*60*60))
	return strftime.Format(cFormat, refTime)
}

func PrintVersion() {
	versionString := "<unknown>"
	buildInfo, ok := debug.ReadBuildInfo()
	if ok {
		const myPath = "github.com/abferm/downtime"
		if buildInfo.Main.Path == myPath {
			versionString = buildInfo.Main.Version
		} else {
			for _, dep := range buildInfo.Deps {
				if dep.Path == myPath {
					versionString = dep.Version
					break
				}
			}
		}
	}

	fmt.Printf("%s go-%s - downtime reporting system\n", os.Args[0], versionString)
	fmt.Println("Copyright (c) 2022 Alex Ferm.\nAll rights reserved.")
	fmt.Println("This software is licensed under the terms and conditions of the MIT License.")

}
