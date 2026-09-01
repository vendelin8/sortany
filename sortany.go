package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var (
	orderBy int
	reverse bool
)

func checkError(descr string, err error, rest ...any) {
	if err != nil {
		fmt.Println("error", descr, err)
		fmt.Println(rest...)
		os.Exit(1)
	}
}

func myUsage() {
	fmt.Printf("Usage: %s [OPTIONS] filename regex\n", os.Args[0])
	flag.PrintDefaults()
}

func init() {
	flag.IntVar(&orderBy, "o", 1, "the index of the column to order by starting with 0")
	flag.BoolVar(&reverse, "r", false, "if the result needs to be inverted")
}

type Result struct {
	Line  []string
	Value float32
}

func main() {
	flag.Usage = myUsage
	flag.Parse()
	if flag.NArg() < 2 {
		flag.Usage()
		os.Exit(1)
	}

	f, err := os.Open(flag.Arg(0))
	checkError("open input file", err)
	defer f.Close()

	scanner := bufio.NewScanner(f)
	re, err := regexp.Compile(flag.Arg(1))
	checkError("compiling regexp", err)
	results := []*Result{}
	for scanner.Scan() {
		l := strings.TrimSpace(scanner.Text())
		if len(l) == 0 {
			continue
		}
		line := re.FindStringSubmatch(l)
		if line == nil {
			fmt.Println("no submatch", l)
			os.Exit(1)
		}
		line = line[1:]
		value, err := strconv.ParseFloat(line[orderBy], 32)
		checkError("regexp running on line", err)
		value32 := float32(value)
		if reverse {
			value32 = -value32
		}
		results = append(results, &Result{line, value32})
	}

	checkError("scanner", scanner.Err())

	sort.Slice(results, func(i, j int) bool {
		return results[i].Value < results[j].Value
	})

	for _, r := range results {
		fmt.Println(strings.Join(r.Line, " "))
	}
}
