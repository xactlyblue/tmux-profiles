package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	"gopkg.in/yaml.v3"
)

type Profile struct {
	Windows []struct {
		Name     *string  `yaml:"name,omitempty"`
		Commands []string `yaml:"commands"`
	} `yaml:"windows"`
}

type Config struct {
	Profiles map[string]Profile `yaml:"profiles"`
}

func execTmuxCommand(args []string) ([]byte, error) {
	var output bytes.Buffer

	cmd := exec.Command("tmux", args...)
	cmd.Stdout = &output
	cmd.Stderr = &output

	if err := cmd.Run(); err != nil {
		if output.Len() > 0 {
			return nil, fmt.Errorf("run command (%v): %v", err, output.String())
		}

		return nil, fmt.Errorf("run command: %v", err)
	}

	return output.Bytes(), nil
}

func attachSession(sessionName string) error {
	cmd := exec.Command("tmux", "attach-session", "-t", sessionName)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("attach session: %v", err)
	}

	return nil
}

func deleteSession(sessionName string) error {
	if _, err := execTmuxCommand([]string{"kill-session", "-t", sessionName}); err != nil {
		return fmt.Errorf("delete session: %v", err)
	}
	return nil
}

func startProfile(name string, data Profile) error {
	return nil
}

func readConfig(filename string) (*Config, error) {
	var config Config

	data, err := os.ReadFile("config.yaml")

	if err != nil {
		return nil, fmt.Errorf("read config: %v", err)
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("unmarshal config: %v", err)
	}

	return &config, nil
}

func main() {
	defer func() {
		if err := deleteSession("testsession"); err != nil {
			fmt.Printf("Error deleting session: %v\n", err)
		} else {
			fmt.Printf("Session 'testsession' deleted successfully.\n")
		}
	}()

	config, err := readConfig("config.yaml")

	if err != nil {
		panic(err)
	}

	for profileName, data := range config.Profiles {
		if err := startProfile(profileName, data); err != nil {
			fmt.Printf("Failed to start profile: %v\n", err)
			os.Exit(1)
		}
	}

	var commands = [][]string{
		{"new-session", "-d", "-s", "testsession"},
		{"send-keys", "-t", "testsession:1", "'nvim'", "Enter"},
	}

	for _, args := range commands {
		if _, err := execTmuxCommand(args); err != nil {
			panic(err)
		}
	}

	if err := attachSession("testsession"); err != nil {
		panic(err)
	}
}
