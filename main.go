package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
)

func main() {
	datesFile, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal("dates-file argument missing")
	}
	defer datesFile.Close()

	upstreamRepo := os.Args[2]
	if err != nil {
		log.Fatal("upstream-repo argument missing")
	} else {
		log.Printf("using %s as upstream git repository", upstreamRepo)
	}

	maxContributionsArg := os.Args[3]
	if err != nil {
		log.Fatal("max-contributions argument missing")
	} else {
		log.Printf("using %s as upstream git repository", upstreamRepo)
	}
	maxContributions, err := strconv.Atoi(maxContributionsArg)
	if err != nil {
		log.Fatalf("converting max-contributions argument failed with %s", err)
	}
	minCommitsPerDay := maxContributions + 1

	repoDir, err := ioutil.TempDir("", "")
	if err != nil {
		log.Fatalf("creating directory for temporary repo failed with %s", err)
	} else {
		log.Printf("using %s as temporary git repository", repoDir)
	}

	cmd := exec.Command("git", "init")
	cmd.Dir = repoDir
	if err := cmd.Run(); err != nil {
		log.Fatalf("git init failed with %s", err)
	}

	cmd = exec.Command("git", "remote", "add", "origin", upstreamRepo)
	cmd.Dir = repoDir
	if err := cmd.Run(); err != nil {
		log.Fatalf("adding origin remote failed with %s", err)
	}

	scanner := bufio.NewScanner(datesFile)
	for scanner.Scan() {
		date := scanner.Text()
		var timestamps []string
		dateParts := strings.Split(date, "/")
		dateRange := strings.Split(dateParts[2], "-")
		if len(dateRange) > 1 {
			first, err := strconv.Atoi(dateRange[0])
			if err != nil {
				log.Fatalf("converting start day of date range failed with %s", err)
			}
			last, err := strconv.Atoi(dateRange[1])
			if err != nil {
				log.Fatalf("converting end day of date range failed with %s", err)
			}
			for i := first; i <= last; i++ {
				timestamps = append(timestamps, fmt.Sprintf("%s-%s-%dT22:13:13", dateParts[0], dateParts[1], i))
			}
		} else {
			timestamps = append(timestamps, fmt.Sprintf("%sT22:13:13", strings.Replace(date, "/", "-", -1)))
		}

		logFilePath := path.Join(repoDir, "knights_of_ni.log")
		logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			log.Printf("creating file %s", logFilePath)
			logFile, err = os.Create(logFilePath)
			if err != nil {
				log.Fatalf("creating file failed with %s", err)
			}
		}
		defer logFile.Close()

		for _, timestamp := range timestamps {
			log.Printf("processing timestamp %s", timestamp)
			for i := 0; i < minCommitsPerDay; i++ {
				_, err := fmt.Fprintln(logFile, "Ni!")
				if err != nil {
					log.Fatalf("writing to file failed with %s", err)
				}

				cmd := exec.Command("git", "add", "knights_of_ni.log")
				cmd.Dir = repoDir
				if err := cmd.Run(); err != nil {
					log.Fatalf("git add failed with %s", err)
				}

				cmd = exec.Command("git", "commit", "-m", `"Ni!"`)
				cmd.Dir = repoDir
				cmd.Env = append(os.Environ(), "GIT_AUTHOR_DATE="+timestamp, "GIT_COMMITTER_DATE="+timestamp)
				if err := cmd.Run(); err != nil {
					log.Fatalf("git commit failed with %s", err)
				}
				fmt.Print(".")
			}
			fmt.Print("\n")
		}

		cmd = exec.Command("git", "push", "-u", "origin", "master")
		cmd.Dir = repoDir
		if err := cmd.Run(); err != nil {
			log.Fatalf("push failed with %s", err)
		}
	}
}
