package main

import (
	"flag"
	"fmt"
)

func main() {
	var folder string
	var email string
	var months int

	flag.StringVar(&folder, "add", "", "add a new folder to scan for Git Repositories")
	flag.StringVar(&email, "email", "your@email.com", "the email to scan")
	flag.IntVar(&months, "months", 6, "number of months back to include")

	flag.Parse()

	if folder != "" {
		scan(folder)
		return
	}

	if email == "" {
		fmt.Println("Please provide an email with --email")
		return
	}

	stats(email, months)
}
