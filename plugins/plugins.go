package plugins

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

type Plugins struct {
	Directory string
	Available map[string]string
}

func New(directory string) (*Plugins, error) {
	files, err := ioutil.ReadDir(directory)
	if err != nil {
		return nil, err
	}

	pl := &Plugins{
		Directory: directory,
		Available: make(map[string]string),
	}

	for _, file := range files {
		log.Debug().Str("command", file.Name()).Msg("found plugin")
		pl.Available[path.Base(file.Name())] = path.Join(directory, file.Name())
	}

	return pl, nil
}

func (p *Plugins) Run(name string, arguments string) (string, error) {
	var buf bytes.Buffer
	if _, ok := p.Available[name]; !ok {
		return "", fmt.Errorf("unknown command: %v", name)
	}
	command := p.Available[name]
	log.Info().Str("command", command).Msg("starting exec")
	split := strings.Split(arguments, " ")
	c := exec.Command(command, split...)
	c.Stdout = &buf
	c.Start()

	done := make(chan error)
	go func() { done <- c.Wait() }()

	timeout := time.After(2 * time.Second)

	select {
	case <-timeout:
		c.Process.Kill()
		log.Error().Msg("command timed out")
	case err := <-done:
		log.Info().Msg(buf.String())
		if err != nil {
			log.Error().Err(err).Msg("command failed")
		}
	}
	return buf.String(), nil
}
