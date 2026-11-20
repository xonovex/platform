package discover

import (
	"errors"
	"path/filepath"
)

// directory is the name of the directory of the agentic configuration in every
// place.
const directory = "agentic"

// ErrNoHome reports a host without a home directory and without the variable
// that would replace it.
var ErrNoHome = errors.New("no home directory")

// Host names the directories of the layers from the facts of a host, so the
// choice is a pure function that an adapter feeds.
type Host struct {
	// GOOS is the operating system, as runtime.GOOS names it.
	GOOS string
	// Getenv returns an environment variable, empty when unset.
	Getenv func(name string) string
	// Home is the home directory of the user, empty when unknown.
	Home string
}

// ManagedDir is /etc/agentic, /Library/Application Support/agentic on macOS
// and %ProgramData%\agentic on Windows.
func (h Host) ManagedDir() string {
	switch h.GOOS {
	case "darwin":
		return filepath.Join("/Library/Application Support", directory)
	case "windows":
		if data := h.Getenv("ProgramData"); data != "" {
			return filepath.Join(data, directory)
		}
		return ""
	default:
		return filepath.Join("/etc", directory)
	}
}

// UserDir is $XDG_CONFIG_HOME/agentic, else %AppData%\agentic on Windows and
// ~/.config/agentic elsewhere.
func (h Host) UserDir() (string, error) {
	return h.place("XDG_CONFIG_HOME", "AppData", ".config")
}

// StateDir is $XDG_STATE_HOME/agentic, else %LocalAppData%\agentic on Windows
// and ~/.local/state/agentic elsewhere.
func (h Host) StateDir() (string, error) {
	return h.place("XDG_STATE_HOME", "LocalAppData", filepath.Join(".local", "state"))
}

// place applies the XDG rule that a relative value is ignored.
func (h Host) place(xdg, windows, home string) (string, error) {
	if base := h.Getenv(xdg); filepath.IsAbs(base) {
		return filepath.Join(base, directory), nil
	}
	if h.GOOS == "windows" {
		if base := h.Getenv(windows); base != "" {
			return filepath.Join(base, directory), nil
		}
	}
	if h.Home == "" {
		return "", ErrNoHome
	}
	return filepath.Join(h.Home, home, directory), nil
}
