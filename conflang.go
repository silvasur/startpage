package main

import (
	"bufio"
	"fmt"
	"io"
)

type toktype int

const (
	tText toktype = iota
	tNextCmd
)

type token struct {
	Type toktype
	Data string
	Line int
}

func scan(r io.Reader, tokens chan<- token, errch chan<- error) {
	emit := func(t toktype, d []byte, line int) {
		if t == tText && len(d) == 0 {
			return
		}
		tokens <- token{t, string(d), line}
	}

	err := func() error {
		br := bufio.NewReader(r)

		escaped := false
		data := []byte{}
		line := 1

		for {
			b, err := br.ReadByte()

			switch err {
			case nil:
			case io.EOF:
				return nil
			default:
				return ErrorAtLine{line, err}
			}

			if b == '\n' {
				line++
			}

			if escaped {
				data = append(data, b)
				escaped = false
				continue
			}

			switch b {
			case '\\':
				escaped = true
			case ' ', '\t':
				emit(tText, data, line)
				data = data[:0]
			case '\n':
				emit(tText, data, line)
				data = data[:0]
				emit(tNextCmd, nil, line)
			default:
				data = append(data, b)
			}
		}

		emit(tText, data, line)
		emit(tNextCmd, nil, line)
		return nil
	}()

	close(tokens)
	errch <- err
}

type command struct {
	Name   string
	Params []string
	Line   int
}

func parse(tokens <-chan token, cmds chan<- command) {
	defer close(cmds)

	startcmd := true
	cmd := command{"", make([]string, 0), 0}

	for tok := range tokens {
		switch tok.Type {
		case tText:
			if startcmd {
				cmd.Name = tok.Data
				cmd.Line = tok.Line
				startcmd = false
			} else {
				cmd.Params = append(cmd.Params, tok.Data)
			}
		case tNextCmd:
			if !startcmd {
				cmds <- cmd
				cmd.Name = ""
				cmd.Params = make([]string, 0)
				startcmd = true
			}
		}
	}
}

type cmdfunc func(params []string) error

var commands = map[string]cmdfunc{
	"nop": func(_ []string) error { return nil },
}

func RegisterCommand(name string, f cmdfunc) {
	commands[name] = f
}

type ErrorAtLine struct {
	Line int
	Err  error
}

func (err ErrorAtLine) Error() string {
	return fmt.Sprintf("%s (at line %d)", err.Err, err.Line)
}

type CommandNotFound string

func (c CommandNotFound) Error() string {
	return fmt.Sprintf("Command \"%s\" not found", c)
}

func RunCommands(r io.Reader) error {
	errch := make(chan error)
	tokens := make(chan token)
	cmds := make(chan command)

	go scan(r, tokens, errch)
	go parse(tokens, cmds)

	for cmd := range cmds {
		f, ok := commands[cmd.Name]
		if !ok {
			return ErrorAtLine{cmd.Line, CommandNotFound(cmd.Name)}
		}

		if err := f(cmd.Params); err != nil {
			return ErrorAtLine{cmd.Line, err}
		}
	}

	return <-errch
}
