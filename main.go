package main

import "github.com/JairoRiver/time_keeper/cmd"

// version is the build version, overridden at link time via
// -ldflags "-X main.version=vX.Y.Z" (see CD-02). Defaults to "dev".
var version = "dev"

func main() {
	cmd.Execute(version)
}
