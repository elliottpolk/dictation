package main

import (
	"bufio"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"gopkg.in/urfave/cli.v2"
)

var version string

var (
	fileFlag = &cli.StringFlag{
		Name:    "dictation-list-file",
		Aliases: []string{"in-file", "in", "f"},
		Value:   "dictation_list.json",
	}

	listFlag = &cli.StringFlag{
		Name:    "list",
		Aliases: []string{"l"},
	}

	delayFlag = &cli.Int64Flag{
		Name:    "delay",
		Aliases: []string{"d"},
		Value:   3,
	}

	interactiveFlag = &cli.BoolFlag{
		Name:    "interactive",
		Aliases: []string{"i"},
		Value:   false,
	}
)

const (
	repeat = "repeat word"
	show   = "show word"
	next   = "next word"
)

type DictList struct {
	Voice string              `json:"voice"`
	Lists map[string][]string `json:"lists"`
}

func main() {
	app := cli.App{
		Version: version,
		Flags:   []cli.Flag{fileFlag, listFlag, delayFlag, interactiveFlag},
		Action:  do,
	}

	app.Run(os.Args)
}

func do(context *cli.Context) error {
	file := context.String(fileFlag.Names()[0])

	in, err := ioutil.ReadFile(file)
	if err != nil {
		return cli.Exit(err, 1)
	}

	dl := &DictList{}
	if err := json.Unmarshal(in, &dl); err != nil {
		return cli.Exit(err, 1)
	}

	interactive := context.Bool(interactiveFlag.Names()[0])
	delay := context.Int64(delayFlag.Names()[0])
	if interactive {
		delay = 0 // no need to have a delay when interactive
	}

	ln := context.String(listFlag.Names()[0])
	list, ok := dl.Lists[ln]
	if !ok {
		return cli.Exit("invalid list name", 1)
	}

	voice := dl.Voice

	for len(list) > 0 {
		clear()

		max := big.NewInt(int64(len(list)))
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return cli.Exit(err, 1)
		}

		index := int(n.Int64())
		word := list[index]

		if err := say(voice, word); err != nil {
			return cli.Exit(err, 1)
		}

		list = append(list[:index], list[index+1:]...)
		if interactive {
			if err := query(voice, word); err != nil {
				return cli.Exit(err, 1)
			}
		}
		time.Sleep(time.Duration(delay) * time.Second)
	}

	return nil
}

// uses the ANSI escape codes
func clear() {
	fmt.Print("\033[H\033[2J")
}

func query(voice, word string) error {
	for {
		opts := []string{show, repeat, next}
		for i, o := range opts {
			fmt.Printf("[%d] %s\n", i+1, o)
		}

		r := bufio.NewReader(os.Stdin)

		in, err := r.ReadString('\n')
		if err != nil {
			return err
		}

		in = strings.TrimSuffix(in, "\n")

		if len(in) < 1 {
			return nil
		}

		i, err := strconv.Atoi(in)
		if err != nil {
			return err
		}

		if i > len(opts) || i < 1 {
			continue // repeat in not a valid selection
		}

		i -= 1 // correct indexing

		res := opts[i]
		switch res {
		case show:
			clear()
			fmt.Printf("%s\n", word)
			time.Sleep(1 * time.Second)

		case repeat:
			clear()
			if err := say(voice, word); err != nil {
				return err
			}

		case next:
			return nil
		}
	}
}

func say(voice, phrase string) error {
	cmd := exec.Command("say", "-v", voice, phrase)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
