package scripts

import (
	"strings"

	"github.com/computeportal/wtsuite/pkg/tokens/patterns"
)

type ControlFileScript struct {
  fName string
	FileScriptData
}

func NewControlFileScript(absPath string, fName string) (*ControlFileScript, error) {
	fileScriptData, err := newFileScriptData(absPath)
	if err != nil {
		return nil, err
	}

	return &ControlFileScript{fName, fileScriptData}, nil
}

func (s *ControlFileScript) Write() (string, error) {
	var b strings.Builder

	// wrap module in a function
	b.WriteString(patterns.NL)
	b.WriteString("function ")
	b.WriteString(s.fName)
	b.WriteString("(){")
	b.WriteString(patterns.NL)

	str, err := s.module.Write(nil, patterns.NL, patterns.TAB)
	if err != nil {
		return "", err
	}

	b.WriteString(str)
	b.WriteString("}")

	return b.String(), nil
}
