package output

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
)

func SendNotifications(updated []string) error {
	updateString := fmt.Sprintf("• %s", strings.Join(updated, "\n •"))
	title := "Nix Package Update!"

	subTitle := "The Following Packages Have Updates"

	var cmd *exec.Cmd
	var cmdName string

	if _, err := exec.LookPath("macos-notifier"); err == nil {
		//macos-notifier is great, but limited...
		cmdName = "macos-notifier"
		comboTitle := fmt.Sprintf("%s\n%s", title, subTitle)
		cmd = exec.Command(cmdName, "--title", comboTitle, "--content", updateString)

	} else {
		cmdName = "osascript"
		n := fmt.Sprintf("display notification \"%s\" with title \"%s\" subtitle \"%s\"", updateString, title, subTitle)
		cmd = exec.Command(cmdName, "-e", n)
	}

	var stderr bytes.Buffer

	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s failed: %w: %s", cmdName, err, stderr.String())
	}

	return nil
}

func GetTable(rows []table.Row) table.Writer {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	// Do a little head | rest style on rows
	t.AppendHeader(rows[0])
	t.AppendRows(rows[1:])
	t.SetStyle(table.StyleRounded)
	t.Style().Options.SeparateRows = true
	return t
}

func FormatUpdatedString(updated []string) string {
	// English is weird
	// SVA = "Subject-verb Agreement"
	// apparently this is how you describe "use singular" or "use plural"
	// TIL
	packageSVA := "Packages"
	if len(updated) == 1 {
		packageSVA = "Package"
	}

	output := fmt.Sprintf("%d Updated %s", len(updated), packageSVA)

	if len(updated) != 0 {
		output = output + fmt.Sprintf(" - %s", strings.Join(updated, ", "))
	}

	return output
}
