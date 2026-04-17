package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"text/template"

	"github.com/google/shlex"

	"gopkg.in/yaml.v3"
)

const (
	defaultConfigPath = "$HOME/.config/tmux-profiles/config.yaml"
)

type Profile struct {
	Windows []struct {
		Name     string   `yaml:"name,omitempty"`
		Commands []string `yaml:"commands"`
	} `yaml:"windows"`
}

type Config struct {
	Profiles map[string]Profile `yaml:"profiles"`
}

type CommandData struct {
	Session string
	Window  string
}

func execCommand(args []string) ([]byte, error) {
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
	if _, err := execCommand([]string{"kill-session", "-t", sessionName}); err != nil {
		return fmt.Errorf("delete session: %v", err)
	}
	return nil
}

func startProfile(name string, data Profile) (err error) {
	defer func() {
		if err == nil {
			return
		}

		if err := deleteSession(name); err != nil {
			fmt.Printf("Failed to automatically delete session: %v\n", err)
		} else {
			fmt.Printf("Automatically deleted session %q due to error\n", name)
		}
	}()

	if _, err := execCommand([]string{"new-session", "-d", "-s", name}); err != nil {
		return fmt.Errorf("new session: %v", err)
	}

	for i, w := range data.Windows {
		var wn string

		if len(w.Name) > 0 {
			wn = w.Name
		} else {
			wn = strconv.Itoa(i + 1)
		}

		if i > 0 {
			if _, err := execCommand([]string{"new-window", "-d", "-t", name, "-n", wn}); err != nil {
				return fmt.Errorf("new window: %v", err)
			}
		} else {
			if _, err := execCommand([]string{"rename-window", "-t", fmt.Sprintf("%s:1", name), wn}); err != nil {
				return fmt.Errorf("rename window: %v", err)
			}
		}

		cd := CommandData{
			Session: name,
			Window:  wn,
		}

		for _, command := range w.Commands {
			args, err := shlex.Split(command)

			if err != nil {
				return fmt.Errorf("parse args: %v", err)
			}

			parsedArgs := make([]string, len(args))

			for i, a := range args {
				tmpl, err := template.New("command").Parse(a)

				if err != nil {
					return fmt.Errorf("new template: %v", err)
				}

				buff := &bytes.Buffer{}

				if err := tmpl.Execute(buff, &cd); err != nil {
					return fmt.Errorf("exec template: %v", err)
				}

				parsedArgs[i] = buff.String()
			}

			if _, err := execCommand(parsedArgs); err != nil {
				return fmt.Errorf("exec command: %v", err)
			}
		}
	}

	if err := attachSession(name); err != nil {
		return fmt.Errorf("attach session")
	}

	return nil
}

func readConfig(filename string) (*Config, error) {
	var config Config

	data, err := os.ReadFile(filename)

	if err != nil {
		return nil, fmt.Errorf("read config: %v", err)
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("unmarshal config: %v", err)
	}

	return &config, nil
}

func main() {
	profileName := os.Args[1]

	path := os.Getenv("TMUX_PROFILES_PATH")

	if len(path) == 0 {
		// TODO: Search multiple paths
		path = defaultConfigPath
	}

	config, err := readConfig(path)

	if err != nil {
		panic(err)
	}

	profile := config.Profiles[profileName]

	if err := startProfile(profileName, profile); err != nil {
		fmt.Printf("Failed to start profile: %v\n", err)
		os.Exit(1)
	}
}
