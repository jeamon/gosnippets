package main

import("time";"os";"sync";"log";"bufio";"strings")

var spamWords []string
var addSpamMutex *sync.RWMutex
var spamLatestStat os.FileInfo

// intialization
func init() {
	
	addSpamMutex = &sync.RWMutex{}
	loadSpamWords() // initial loading of spam words from file
}

// load spam wordslist and save
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
    addSpamMutex.Lock() // On verrou la liste pour manipulation 
    // as needed to reflect same state as file - clean and recreate the slice
    spamWords = nil
	spamWords = []string{}
	for scanner.Scan() {
	    spam = strings.Trim(strings.TrimSpace(scanner.Text()),"\r\n")
	    if spam != "" {
	    	spamWords = append(spamWords, spam)
    	}
    }
    log.Println(spamWords) // display for checking
    addSpamMutex.Unlock()
}

// check every interval hour and update if changes
func updateSpamWords(interval int) {
	for {
		stat, err := os.Stat("config/spam-words.txt")
		if err != nil {
			log.Println("[ Eror ] Failed to get statistics of spam words file. ErrMsg -", err)
		} else {
			if stat.Size() != spamLatestStat.Size() || stat.ModTime() != spamLatestStat.ModTime() {
				loadSpamWords() // load when size or latest modification time changed
			}
		}
		time.Sleep(time.Duration(interval) * time.Hour)
	}
}


// change this file config/spam-words.txt content and observe
func main() {
	doneChan := make(chan bool)
	go updateSpamWords(1) // check to update each 1 hour if any changes
	<- doneChan
}


/* 

type Message struct { 
	FullName string
	Email string
	Subject string
	Content string
	Errors map[string]string
}

check if subject or contact is suspicious
func (msg *Message) isSpamMessage() bool {
	// acquire the lock and ensure its release
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

// snipped to insert into the contact handler and send fake confirmation for spam message
if msg.isSpamMessage() {
	// routine to handle goes here
	// you can ignore user message or send fake confirmation
}
*/