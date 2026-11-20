package cmd

import "testing"

func TestNewRootCommandRegistersSubcommandsAndFlags(t *testing.T) {
	root := NewRootCommand(testLog)

	for _, name := range []string{"run", "completion"} {
		if command, _, err := root.Find([]string{name}); err != nil || command.Name() != name {
			t.Errorf("Find(%q) = (%v, %v), want the %s command", name, command, err, name)
		}
	}
	run, _, _ := root.Find([]string{"run"})
	for _, flag := range []string{"agent", "isolation", "provision", "network", "require-egress-restricted", "dry-run"} {
		if run.Flags().Lookup(flag) == nil {
			t.Errorf("run command has no --%s flag", flag)
		}
	}
	if root.PersistentFlags().Lookup("verbose") == nil {
		t.Error("root command has no --verbose flag")
	}
}
