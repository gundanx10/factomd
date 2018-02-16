package interpreter

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	. "github.com/FactomProject/electiontesting/interpreter/common"
	. "github.com/FactomProject/electiontesting/interpreter/dictionary"
	. "github.com/FactomProject/electiontesting/interpreter/names"
	. "github.com/FactomProject/electiontesting/interpreter/stack"
)

type Interpreter struct {
	Stack     // Data stack is integral
	C         Stack
	Compiling int
	DictStack []Dictionary
	Input     *bufio.Reader
	Line      string
	NameManager
}

func NewInterpreter() Interpreter {
	var i Interpreter
	i.Stack = NewStack()
	i.C = NewStack()
	i.DictStack = make([]Dictionary, 0)
	i.NameManager = NewNameManager()
	return i
}

// Convert a string to a name
func (i *Interpreter) Lookup(s string) interface{} {
	n := i.GetName(s)
	return n
}

// Push a dictionary on the stack
func (i *Interpreter) DictionaryPush(d Dictionary) {
	i.DictStack = append([]Dictionary{d}, i.DictStack...)
}
func (i *Interpreter) DictionaryPop() { i.DictStack = i.DictStack[1:] }

func (i *Interpreter) Exec3(x interface{}) {
	var flags FlagsStruct // assume its not immediate and not executable

	//	fmt.Printf("Exec3(%v) ", x)
	//	i.PStack()

	// check for thing with no flags and create flags for them
	f, ok := x.(func()) // is it a raw Go Function? Then it's executable but not immediate
	if ok {
		flags.Immediate = false
		flags.Executable = true
	} else {
		immediateFunc, ok := x.(ImmediateFunc) // Should not have to manually check this!!!
		if ok {
			immediateFunc.Func()
			return
		}
	}

	_, ok = x.(HasFlags)
	if ok {
		flags = x.(HasFlags).GetFlags()
	}

	if flags.Immediate || (flags.Executable && i.Compiling == 0) {
		// Got an executable thing
		switch x.(type) {
		case Array:
			for _, y := range x.(Array).Data {
				switch y.(type) {
				case Array:
					i.Push(y)
				default:
					i.Exec3(y)
				}
			} // for all elements of the executable array
		case func():
			f() // execute the primitive
		case Name:
			// find the name in the dict stack
			for _, d := range i.DictStack {
				e, ok := d[x.(Name)]
				if ok {
					i.Exec3(e)
				}
			}
			panic("Undefined " + i.GetString(x.(Name)))

		default:
			i.Push(x) // Maybe should panic here but ...
		} // switch on type

	} else {
		i.Push(x)
	}

}

// execute one thing
func (i *Interpreter) InterpretString(s string) {
	fmt.Printf("Exec2(\"%s\")\n", s)
	if ii, err := strconv.Atoi(s); err == nil {
		i.Exec3(ii)
	} else if b, err := strconv.ParseBool(s); err == nil {
		i.Exec3(b)
	} else if f, err := strconv.ParseFloat(s, 64); err == nil {
		i.Exec3(f)
	} else if ii, err := strconv.ParseInt(s, 10, 64); err == nil {
		i.Exec3(ii)
	} else if u, err := strconv.ParseUint(s, 10, 64); err == nil {
		i.Exec3(u)
	} else {
		// Wasn't a literal
		e := i.Lookup(s)
		i.Exec3(e)
	}
}

func (i *Interpreter) InterpretLine(line string) {
	//	fmt.Printf("Interpret(\"%s\")\n", line)
	defer func() { i.Line = i.Line }()
	i.Line = line

	var s string
	for {
		// Scan a string from the current line (possible modified by execution)
		line := i.Line
		line = strings.TrimSpace(line)
		n, err := fmt.Sscan(line, &s)
		if n == 1 {
			line = line[len(s):] // Trim off the string and the ws following
			i.Line = line
			if s != "" {
				i.InterpretString(s) // execute the string
			}
		}
		if i.Line == "" {
			break
		}
		if err == io.EOF {
			return
		}
		if err != nil {
			panic(err)
		}
	} // Until we have done all the strings on the line
} // till EOF or error

func (i *Interpreter) Interpret(source io.Reader) {
	defer func() { i.Input = i.Input }() // Reset i.Input when we exit
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Error:", r)
		}
	}()

	i.Input = bufio.NewReader(source) // save the source off for primitives that need it
	for {
		var line string
		for {
			chunk, isPrefix, err := i.Input.ReadLine()
			line += string(chunk) // append this piece of the line
			if err == io.EOF || line == "" {
				return
			}
			if err != nil {
				panic(err)
			}
			if !isPrefix {
				break
			}
		} // Until we get a whole line
		i.InterpretLine(line)
	}
}
