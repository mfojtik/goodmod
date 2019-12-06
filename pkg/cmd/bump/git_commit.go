package bump

import (
	"fmt"
	"os/exec"
	"strings"
)

func commitGoMod() error {
	if out, err := exec.Command("git", "add", "go.mod").CombinedOutput(); err != nil {
		return fmt.Errorf("%s", out)
	}
	if out, err := exec.Command("git", "commit", "-m", "bump(*): go.mod changes").CombinedOutput(); err != nil {
		return fmt.Errorf("%s", out)

	}
	return nil
}

func commitVendor(commits []string) error {
	messages := []string{"bump(*): go mod vendor", ""}
	messages = append(messages, commits...)

	if out, err := exec.Command("go", "mod", "tidy").CombinedOutput(); err != nil {
		return fmt.Errorf("%s", out)
	}
	if out, err := exec.Command("go", "mod", "vendor").CombinedOutput(); err != nil {
		return fmt.Errorf("%s", out)
	}
	if out, err := exec.Command("git", "add", "go.sum", "./vendor").CombinedOutput(); err != nil {
		return fmt.Errorf("%s", out)
	}
	if out, err := exec.Command("git", "commit", "-m", fmt.Sprintf("%s", strings.Join(messages, "\n"))).CombinedOutput(); err != nil {
		return fmt.Errorf("%s", out)
	}
	return nil
}
