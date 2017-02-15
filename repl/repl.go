package repl

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/signal"

	"github.com/goruby/goruby/ast"
	"github.com/goruby/goruby/evaluator"
	"github.com/goruby/goruby/lexer"
	"github.com/goruby/goruby/object"
	"github.com/goruby/goruby/parser"
)

const PROMPT = "girb:%03d> "

func Start(in io.Reader, out io.Writer) {
	printChan := make(chan string)
	sigChan := make(chan os.Signal, 4)
	signal.Notify(sigChan, os.Interrupt, os.Kill)
	go loop(in, printChan)
	for {
		select {
		case evaluated := <-printChan:
			fmt.Fprintf(out, "%s", evaluated)
		case <-sigChan:
			fmt.Fprintln(out)
			return
		}
	}
}

func loop(in io.Reader, out chan<- string) {
	scanner := bufio.NewScanner(in)
	counter := 1
	env := object.NewEnvironment()
	for {
		out <- fmt.Sprintf(PROMPT, counter)
		counter++
		scanned := scanner.Scan()
		if !scanned {
			out <- fmt.Sprintln()
			return
		}

		line := scanner.Text()
		node, err := parseLine(line)
		if err != nil {
			out <- fmt.Sprintf("%s", err.Error())
			continue
		}

		evaluated := evaluator.Eval(node, env)
		if evaluated != nil {
			out <- fmt.Sprintf("=> %s\n", evaluated.Inspect())
		}
	}
}

func parseLine(line string) (ast.Node, error) {
	l := lexer.New(line)
	p := parser.New(l)
	var err error
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		err = mergeParserErrors(p.Errors())
	}
	return program, err
}

func mergeParserErrors(errors []error) error {
	var buf bytes.Buffer
	printParserErrors(&buf, errors)
	return fmt.Errorf(buf.String())
}

func printParserErrors(out io.Writer, errors []error) {
	fmt.Println("Parser errors: ")
	for _, err := range errors {
		fmt.Fprintf(out, "\t%s\n", err.Error())
	}
}
