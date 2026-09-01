package main

import (
	"bufio"
	"flag"
	"fmt"
	"math"
	"os"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var (
	collects = map[string]*Collect{}
	results  = []*Result{}
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

type Collect []float32

type Result struct {
	Name   string
	Count  int
	Avg    float32
	StdDev float32
}

type Local struct {
	Name  string
	Value float32
}

func main() {
	f, err := os.Open(os.Args[1])
	checkError("open input file", err)
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		l := strings.TrimSpace(scanner.Text())
		f := flag.NewFlagSet("sortany", flag.ExitOnError)
		orderBy := f.Int("o", 1, "the index of the column to order by starting with 0")
		reverse := f.Bool("r", false, "if the result needs to be inverted")
		f.Usage = myUsage
		ls := strings.Split(l, ";")
		var b strings.Builder
		b.WriteString(path.Dir(os.Args[1]))
		b.WriteString("/b")
		b.WriteString(ls[len(ls)-2])
		ls[len(ls)-2] = b.String()
		f.Parse(ls)
		args := f.Args()
		readFile(f.Arg(0), args[len(args)-1], *reverse, *orderBy)
	}
	checkError("scanner", scanner.Err())

	for name, collect := range collects {
		cl := len(*collect)
		var sum float32
		for _, c := range *collect {
			sum += c
		}
		avg := sum / float32(cl)
		sum = 0
		for _, c := range *collect {
			tmp := avg - c
			sum += tmp * tmp
		}
		stdDev := float32(math.Sqrt(float64(sum / float32(cl-1))))
		results = append(results, &Result{name, cl, avg, stdDev})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Avg < results[j].Avg
	})
	for _, r := range results {
		fmt.Println(r.Name, r.Avg, r.Count, r.StdDev)
	}
}

func readFile(filename, reStr string, reverse bool, orderBy int) {
	f, err := os.Open(filename)
	checkError("open input file", err)
	defer f.Close()

	scanner := bufio.NewScanner(f)
	re, err := regexp.Compile(reStr)
	checkError("compiling regexp", err)
	locals := []*Local{}
	for scanner.Scan() {
		l := strings.TrimSpace(scanner.Text())
		if len(l) == 0 {
			continue
		}
		line := re.FindStringSubmatch(l)
		if line == nil {
			fmt.Println("no submatch", l, reStr)
			os.Exit(1)
		}
		line = line[1:]
		name := line[0]
		if _, ok := collects[name]; !ok {
			collects[name] = &Collect{}
		}
		value, err := strconv.ParseFloat(line[orderBy], 32)
		checkError("regexp running on line", err)
		value32 := float32(value)
		if reverse {
			value32 = -value32
		}
		locals = append(locals, &Local{name, value32})
	}

	checkError("scanner", scanner.Err())

	sort.Slice(locals, func(i, j int) bool {
		return locals[i].Value < locals[j].Value
	})

	lastValue := locals[0].Value
	index := 0
	for i, l := range locals {
		if l.Value > lastValue {
			index++
		}
		locals[i].Value = float32(index)
	}

	mult := 1.0 / float32(index)
	for _, l := range locals {
		*collects[l.Name] = append(*collects[l.Name], l.Value*mult)
	}
}
