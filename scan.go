package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/user"
	"strings"
)

// getDotFilePath defines a fixed location to store repo data (~/.gogitlocalstats)
func getDotFilePath() string {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	dotFile := usr.HomeDir + "/.gogitlocalstats"
	return dotFile
}

// openFile opens the file located at filepath. Creates it if it does not exist.
func openFile(filePath string) *os.File {
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		if os.IsNotExist(err) {
			// File does not exist → create it
			f, err = os.Create(filePath)
			if err != nil {
				panic(err)
			}
		} else {
			// Other error
			panic(err)
		}
	}
	return f
}

// parseFileLinesToSlice reads a file line by line and stores each line in a slice of strings
func parseFileLinesToSlice(filePath string) []string {
	f := openFile(filePath)
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		if err != io.EOF {
			panic(err)
		}
	}
	return lines
}

// sliceContains returns true if slice contains value
func sliceContains(slice []string, value string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

// joinSlices adds elements from the 'new' slice into 'existing', avoiding duplicates
func joinSlices(new []string, existing []string) []string {
	for _, i := range new {
		if !sliceContains(existing, i) {
			existing = append(existing, i)
		}
	}
	return existing
}

// dumpStringsSliceToFile writes the content of a string slice to a file (overwriting existing content)
func dumpStringsSliceToFile(repos []string, filePath string) {
	content := strings.Join(repos, "\n")
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		log.Fatal(err)
	}
}

// addNewSliceElementsToFile merges new repo paths with the ones already stored in file,
// removes duplicates, and saves the updated list back to the file.
func addNewSliceElementsToFile(filePath string, newRepos []string) {
	existingRepos := parseFileLinesToSlice(filePath) // File → Slice
	repos := joinSlices(newRepos, existingRepos)     // Merge without duplicates
	dumpStringsSliceToFile(repos, filePath)          // Slice → File
}

// recursiveScanFolder starts a recursive search for Git repositories in the folder subtree
func recursiveScanFolder(folder string) []string {
	return scanGitFolders(make([]string, 0), folder)
}

// scan scans a folder for Git repositories and writes results to the dotfile
func scan(folder string) {
	fmt.Printf("Found folders:\n\n")

	repositories := recursiveScanFolder(folder) // Get slice of repo paths
	filePath := getDotFilePath()                // Path of dotfile
	addNewSliceElementsToFile(filePath, repositories)

	fmt.Printf("\n\nSuccessfully added\n\n")
}

// scanGitFolders recursively searches a folder for Git repositories (.git folders)
// Returns a slice of repo root paths, skipping vendor and node_modules directories
func scanGitFolders(folders []string, folder string) []string {
	// Remove trailing slash
	folder = strings.TrimSuffix(folder, "/")

	f, err := os.Open(folder)
	if err != nil {
		log.Fatal(err)
	}
	files, err := f.Readdir(-1) // Read all directory entries (files + subdirectories)
	f.Close()
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		if file.IsDir() {
			path := folder + "/" + file.Name()
			if file.Name() == ".git" {
				// Found a git repo → store its parent folder
				repoPath := strings.TrimSuffix(path, "/.git")
				fmt.Println(repoPath)
				folders = append(folders, repoPath)
				continue
			}
			if file.Name() == "vendor" || file.Name() == "node_modules" {
				continue // Skip dependencies folders
			}
			// Recurse into subfolder
			folders = scanGitFolders(folders, path)
		}
	}

	return folders
}
/*
// main function for testing
func main() {
	// Example: scan your home directory (or any folder you want)
	scan(os.Getenv("HOME"))
}
*/
