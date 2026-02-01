package logging

import "flag"

func init() {
	flag.String("loglevel", "string", "log level to use")
}
