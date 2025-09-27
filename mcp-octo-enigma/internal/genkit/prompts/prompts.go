package prompts

import (
	"os"
)

func LoadDotprompt(path string) (string, error) { b, err := os.ReadFile(path); if err != nil { return "", err }; return string(b), nil }
