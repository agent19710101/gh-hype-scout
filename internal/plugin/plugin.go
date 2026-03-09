package plugin

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"strings"

	"github.com/agent19710101/gh-hype-scout/internal/output"
)

type Processor interface {
	Process(report output.DeltaReport) error
}

type ExternalCommand struct {
	Command string
}

func (p ExternalCommand) Process(report output.DeltaReport) error {
	if strings.TrimSpace(p.Command) == "" {
		return nil
	}
	payload, err := json.Marshal(report)
	if err != nil {
		return err
	}
	cmd := exec.Command("bash", "-lc", p.Command)
	cmd.Stdin = bytes.NewReader(payload)
	return cmd.Run()
}
