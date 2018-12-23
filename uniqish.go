package main

import (
    "bufio"
    "fmt"
    "os"
)

func main() {
    counts := make(map[string]int)
    dates := make(map[string]string)
    input := bufio.NewScanner(os.Stdin)
    for input.Scan() {
        counts[input.Text()[15:]]++
        dates[input.Text()[15:]] = input.Text()[:15]
    }
    for line, n := range counts {
        fmt.Printf("%d\t%s\t%s\n", n, dates[line], line)
    }
}
