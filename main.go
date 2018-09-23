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
	"strings"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/urfave/cli.v2"
)

var version string

type DictList struct {
	Voice string              `json:"voice"`
	Lists map[string][]string `json:"lists"`
}

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

	practiceCommand = &cli.Command{
		Name:    "practice",
		Aliases: []string{"study", "p"},
		Flags: []cli.Flag{
			fileFlag,
			listFlag,
		},
		Action: practice,
	}

	quizCommand = &cli.Command{
		Name:    "quiz",
		Aliases: []string{"test", "q"},
		Flags: []cli.Flag{
			fileFlag,
			listFlag,
			delayFlag,
		},
		Action: quiz,
	}
)

func main() {
	app := cli.App{
		Version: version,
		Commands: []*cli.Command{
			practiceCommand,
			quizCommand,
		},
	}

	app.Run(os.Args)
}

func cls() {
	fmt.Print("\033[H\033[2J")
}

func cln() {
	fmt.Print("\033[1A\033[K")
}

// currently tested on macOS only
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

func practice(context *cli.Context) error {
	cls()

	dl, err := parse(context.String(fileFlag.Names()[0]))
	if err != nil {
		return cli.Exit(err, 1)
	}

	ln := context.String(listFlag.Names()[0])
	list, ok := dl.Lists[ln]
	if !ok {
		return cli.Exit("invalid list name", 1)
	}

	voice := dl.Voice

	for len(list) > 0 {
		word := list[0]
		list = list[1:] // remove word from list

		fmt.Println(word)
		if err := say(voice, word); err != nil {
			return cli.Exit(err, 1)
		}

		for {
			fmt.Print("hit enter to continue or [repeat|again] to hear the word again: ")
			in, err := bufio.NewReader(os.Stdin).ReadString('\n')
			if err != nil {
				return cli.Exit(err, 1)
			}

			in = strings.TrimSpace(strings.TrimSuffix(in, "\n"))
			if len(in) < 1 {
				cln()
				break
			}

			cln()
			if in == "repeat" || in == "again" {
				if err := say(voice, word); err != nil {
					return cli.Exit(err, 1)
				}
			}
		}
	}

	return nil
}

func quiz(context *cli.Context) error {
	cls()

	dl, err := parse(context.String(fileFlag.Names()[0]))
	if err != nil {
		return cli.Exit(err, 1)
	}

	ln := context.String(listFlag.Names()[0])
	list, ok := dl.Lists[ln]
	if !ok {
		return cli.Exit("invalid list name", 1)
	}

	delay := context.Int64(delayFlag.Names()[0])
	voice := dl.Voice

	for len(list) > 0 {
		max := big.NewInt(int64(len(list)))
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return cli.Exit(err, 1)
		}

		index := int(n.Int64())
		word := list[index]

		for i := 0; i < 3; i++ {
			fmt.Print(".")
			if err := say(voice, word); err != nil {
				return cli.Exit(err, 1)
			}
			time.Sleep(time.Duration(delay) * time.Second)
		}
		fmt.Println(" ", word)

		time.Sleep(200 * time.Millisecond) // add a slight delay to reduce the output "shock"

		fmt.Print("enter to continue: ")
		if _, err := bufio.NewReader(os.Stdin).ReadString('\n'); err != nil {
			return cli.Exit(err, 1)
		}
		cln()
		list = append(list[:index], list[index+1:]...)
	}

	return nil
}

func parse(file string) (*DictList, error) {
	in, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read in file")
	}

	dl := &DictList{}
	if err := json.Unmarshal(in, &dl); err != nil {
		return nil, errors.Wrap(err, "unable to parse file")
	}

	return dl, nil
}
