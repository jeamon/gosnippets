package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

var spamWords []string
var addSpamMutex *sync.RWMutex
var spamLatestStat os.FileInfo

func init() {
	// initialize global spamWords list RW mutex.
	addSpamMutex = &sync.RWMutex{}
	// always load spam words from file at startup.
	loadSpamWords()
}

// loadSpamWords read the content of the file "config/spam-words.txt" and add each line to the
// spam words list. It also update the spamLatestStat variable with the file attributes.
func loadSpamWords() {
	var spam string
	fichier, err := os.Open("config/spam-words.txt")
	if err != nil {
		log.Println("[ Eror ] Failed to load spam words file. ErrMsg -", err)
		os.Exit(1)
	}

	defer fichier.Close()

	spamLatestStat, err = fichier.Stat()
	if err != nil {
		log.Println("[ Eror ] Failed to get latest statistics of spam words file. ErrMsg -", err)
	}

	scanner := bufio.NewScanner(fichier)
	// lock the list for processing.
	addSpamMutex.Lock()
	// as needed to reflect same state as file - clean and recreate the slice.
	spamWords = nil
	spamWords = []string{}
	// loop over the file content and build the spam words list.
	for scanner.Scan() {
		spam = strings.Trim(strings.TrimSpace(scanner.Text()), "\r\n")
		if spam != "" {
			spamWords = append(spamWords, spam)
		}
	}
	// release the lock.
	addSpamMutex.Unlock()
	// display for checking if needed.
	log.Println(spamWords)
}

// updateSpamWords check every interval hour and update spam words list in case the file changed.
func updateSpamWords(interval int) {

	for {
		stat, err := os.Stat("config/spam-words.txt")
		if err != nil {
			log.Println("[ Eror ] Failed to get statistics of spam words file. ErrMsg -", err)
		} else {
			if stat.Size() != spamLatestStat.Size() || stat.ModTime() != spamLatestStat.ModTime() {
				// size or latest modification time changed so load file content.
				loadSpamWords()
			}
		}
		// wait until next interval hour(s).
		time.Sleep(time.Duration(interval) * time.Hour)
	}
}

// in case you want to experiment this program change this file
// "config/spam-words.txt" content and observe the output list.
func main() {

	// every 1 hour check for any changes and updates if any.
	go updateSpamWords(1)

	fmt.Println("\nPress the Enter Key to leave the program.")
	// wait for Enter Key.
	fmt.Scanln()
}

// below is a short demo of how to use this above routine into your contact message handler.
/*

// contact submission message format.
type Message struct {
	FullName string
	Email string
	Subject string
	Content string
	Errors map[string]string
}

// check if subject or content is suspicious.
func (msg *Message) isSpamMessage() bool {
	// acquire the lock and ensure its release.
	addSpamMutex.RLock()
	defer addSpamMutex.RUnlock()
	// loop over the spam words list and stop once there is a hit
	for _, spam := range spamWords {
		if strings.Contains(msg.Content, spam) || strings.Contains(msg.Subject, spam) {
			return true
		}
	}

	return false
}

// snipped to insert into the contact handler and send fake confirmation for spam message.
if msg.isSpamMessage() {
	// routine to handle goes here
	// you can ignore user message or send fake confirmation
}
*/
