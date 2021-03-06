/*
 * gocli
 * Copyright (c) 2015 Yieldbot, Inc.
 * For the full copyright and license information, please view the LICENSE.txt file.
 */

// Package gocli is a CLI library that provides subcommand handling, tidy usage and version printing.
package gocli

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"sort"
	"strings"
)

// Cli represent command line interface
type Cli struct {
	// Name is the cli name
	Name string

	// Version is the cli version
	Version string

	// Description is the cli description
	Description string

	// Commands contains the subcommand list of the cli
	Commands map[string]string

	// SubCommand contains the runtime subcommand
	SubCommand string

	// SubCommandArgs contains the args of the runtime subcommand
	SubCommandArgs []string

	// SubCommandArgsMap contains the args of the runtime subcommand as mapped
	SubCommandArgsMap map[string]string

	// Flags contains flags
	Flags map[string]string

	// LogOut is logger for stdout
	LogOut *log.Logger

	// LogErr is logger for stderr
	LogErr *log.Logger
}

// Init initializes Cli instance
func (cl *Cli) Init() {

	// Init flag
	if !flag.Parsed() {
		flag.Parse()
	}

	// Init loggers
	cl.LogOut = log.New(os.Stdout, "", log.LstdFlags)
	cl.LogErr = log.New(os.Stderr, "", log.LstdFlags)

	// Init flags
	cl.Flags = make(map[string]string)
	flag.VisitAll(func(f *flag.Flag) {
		cl.Flags[f.Name] = f.Value.String()
	})

	// Init args
	if len(os.Args) > 1 {

		// Iterate the args
		for _, arg := range os.Args {
			// If the arg is in command list then
			if _, ok := cl.Commands[arg]; ok {
				cl.SubCommand = arg // set as command
			} else {
				// Otherwise add it to subcommand args
				if cl.SubCommand != "" {
					cl.SubCommandArgs = append(cl.SubCommandArgs, arg)
				}
			}
		}

		// Init subcommand args map
		cl.SubCommandArgsMap = make(map[string]string)
		var curArg string
		for _, v := range cl.SubCommandArgs {
			// If it's an arg then
			if strings.HasPrefix(v, "-") {
				curArg = strings.TrimLeft(v, "-")
				if len(curArg) > 0 {
					cl.SubCommandArgsMap[curArg] = ""
				}
			} else {
				// Otherwise add it to current arg or add it as arg
				if len(curArg) > 0 {
					cl.SubCommandArgsMap[curArg] = v
					curArg = ""
				} else {
					cl.SubCommandArgsMap[v] = ""
				}
			}
		}
	}
}

// PrintVersion prints version information
func (cl Cli) PrintVersion(extra bool) {
	var ver string

	if extra == true {
		ver += fmt.Sprintf("Bin Version : %s\n", strings.TrimPrefix(cl.Version, "v"))
		ver += fmt.Sprintf("Go version  : %s", runtime.Version())
	} else {
		ver = fmt.Sprintf("%s", strings.TrimPrefix(cl.Version, "v"))
	}

	fmt.Println(ver)
}

// PrintUsage prints usage info
// Usage format follows common convention for Go apps
func (cl Cli) PrintUsage() {

	// Init vars
	type flagInfo struct {
		nameu    string
		name     string
		usage    string
		defValue string
	}

	// Find the longest command for alignment
	maxlen := 0
	if len(cl.Commands) > 0 {
		for c := range cl.Commands {
			if len(c) > maxlen {
				maxlen = len(c)
			}
		}
	}

	// Iterate flags
	flagList := make(map[string]*flagInfo)
	flag.VisitAll(func(f *flag.Flag) {

		// If the flag name starts with `test.` then
		if strings.Index(f.Name, "test.") == 0 {
			return
		}

		// Set key by the flag usage for grouping
		key := fmt.Sprint(f.Usage)

		// Init usage name
		nameu := "-" + f.Name
		if len(f.Name) > 2 {
			nameu = "-" + nameu
		}

		// If the flag exists then
		if _, ok := flagList[key]; ok {
			// Merge names
			flagList[key].nameu += ", " + nameu
		} else {
			// Otherwise add the flag
			flagList[key] = &flagInfo{
				nameu:    nameu,
				name:     f.Name,
				usage:    f.Usage,
				defValue: f.DefValue,
			}
		}

		// Check and set maximum length for alignment
		if len(flagList[key].nameu) > maxlen {
			maxlen = len(flagList[key].nameu)
		}
	})

	var maxlenF = fmt.Sprintf("%d", maxlen)

	// Fixed flag list
	flagListF := []string{}
	for _, v := range flagList {
		flagline := fmt.Sprintf("%-"+maxlenF+"s : %s", v.nameu, v.usage)
		if v.defValue != "false" && v.defValue != "" {
			flagline += " (default \"" + v.defValue + "\")"
		}
		flagListF = append(flagListF, flagline)
	}
	sort.Strings(flagListF)

	// Fixed command list
	cmdListF := []string{}
	for cn, cv := range cl.Commands {
		cmdListF = append(cmdListF, fmt.Sprintf("%-"+maxlenF+"s : %s", cn, cv))
	}
	sort.Strings(cmdListF)

	// Header and description
	usage := "Usage: " + cl.Name + " [OPTIONS] COMMAND [arg...]\n\n"
	if cl.Description != "" {
		usage += cl.Description + "\n\n"
	}

	// Options
	if len(flagListF) > 0 {
		usage += "Options:\n"
		for _, f := range flagListF {
			usage += fmt.Sprintf("  %s\n", f)
		}
	}

	// Commands
	if len(cmdListF) > 0 {
		usage += "\nCommands:\n"
		for _, c := range cmdListF {
			usage += fmt.Sprintf("  %s\n", c)
		}
	}

	fmt.Println(usage)
}

// Table represent tabular data as a table
type Table struct {
	data     [][]string
	colSizes map[int]int
}

// Data gets data
func (t *Table) Data() [][]string {
	return t.data
}

// SetData sets a data by the given row, column and value
func (t *Table) SetData(row, col int, val string) error {

	// Check row and column numbers
	if row < 1 || col < 1 {
		return errors.New("invalid row or column index")
	}

	// Increase the row capacity if it's necessary
	if row > len(t.data) {
		nt := make([][]string, row)
		copy(nt, t.data)
		t.data = nt
	}

	// Increase the column capacity if it's necessary
	if col > len(t.data[row-1]) {
		nr := make([]string, col)
		copy(nr, t.data[row-1])
		t.data[row-1] = nr
	}

	// Set the value
	t.data[row-1][col-1] = val

	// Set the column size for alignment
	if t.colSizes == nil {
		t.colSizes = make(map[int]int)
	}

	if len(val) > t.colSizes[col-1] {
		t.colSizes[col-1] = len(val)
	}

	return nil
}

// AddRow adds a row data by the given row number and column values
func (t *Table) AddRow(row int, cols ...string) error {

	// Iterate rows and set data
	for i, v := range cols {
		if err := t.SetData(row, i+1, v); err != nil {
			return err
		}
	}
	return nil
}

// PrintData prints data
func (t *Table) PrintData() {

	if len(t.data) == 0 {
		return
	}

	// Print data
	var rowVal string
	var colSize string
	for _, row := range t.data {
		rowVal = ""
		for i, c := range row {
			colSize = fmt.Sprintf("%d", t.colSizes[i])
			rowVal += fmt.Sprintf("%-"+colSize+"s\t", c)
		}
		fmt.Println(rowVal)
	}
}
