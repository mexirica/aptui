package platform

import (
	"os"
	"os/exec"
)

// OnTermux is true when running inside Termux, where sudo is unavailable.
var OnTermux = os.Getenv("TERMUX_VERSION") != "" || os.Getenv("TERMUX_API_VERSION") != ""

// SudoCmd builds an exec.Cmd, prepending "sudo" unless running on Termux.
func SudoCmd(name string, args ...string) *exec.Cmd {
	if OnTermux {
		return exec.Command(name, args...)
	}
	return exec.Command("sudo", append([]string{name}, args...)...)
}

// SudoCmdSilent builds a non-interactive exec.Cmd (sudo -n), or a plain
// command on Termux.
func SudoCmdSilent(name string, args ...string) *exec.Cmd {
	if OnTermux {
		return exec.Command(name, args...)
	}
	return exec.Command("sudo", append([]string{"-n", name}, args...)...)
}

// AptPath returns the base APT configuration path. On Termux this is
// $PREFIX/etc/apt; on standard Linux it is /etc/apt.
func AptPath(subpath string) string {
	if OnTermux {
		prefix := os.Getenv("PREFIX")
		if prefix == "" {
			prefix = "/data/data/com.termux/files/usr"
		}
		return filepath.Join(prefix, "etc", "apt", subpath)
	}
	return filepath.Join("/etc/apt", subpath)
}
