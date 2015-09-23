package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/codegangsta/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "songsearch"
	app.Usage = "Search for the occurences of a word in a database of lyrics"
	app.Action = func(c *cli.Context) {
		fmt.Println("Please specify a command")
	}

	app.Commands = []cli.Command{
		{
			Name:    "search",
			Aliases: []string{"s"},
			Usage:   "search [filePath]",
			Action: func(c *cli.Context) {
				readFile(c)
			},
		},
	}

	app.Run(os.Args)
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

type song struct {
	title  string
	artist string
	words  []string
	id     int
}

type word struct {
	word      string
	songID    int
	positions []int
}
type words []word

func (w words) Len() int {
	return len(w)
}
func (w words) Swap(i, j int) {
	w[i], w[j] = w[j], w[i]
}
func (w words) Less(i, j int) bool {
	return len(w[i].positions) > len(w[j].positions)
}

func readFile(c *cli.Context) {
	begin := time.Now()
	var filePath string
	if len(c.Args()) == 1 {
		filePath = c.Args()[0]
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			fmt.Println("No such file or directory:", filePath)
			os.Exit(1)
		}
	} else {
		fmt.Println("You must specify a filepath as the first and only argument")
		os.Exit(1)
	}
	fmt.Println(filePath)
	f, err := os.Open(filePath)
	defer f.Close()
	check(err)
	scanner := bufio.NewScanner(f)

	onTitle, onArtist := true, false
	currID := 0
	currSong := song{id: currID}
	songChan := make(chan song, 3)
	tableChan := make(chan map[string]words)
	var songs []song
	go processSong(songChan, tableChan)
	fmt.Println("Starting Annexation")
	for scanner.Scan() {
		// Go line by line through the input file
		line := scanner.Text()
		if onTitle {
			currSong.title = line
			onTitle = false
			onArtist = true
		} else if onArtist {
			currSong.artist = line
			onArtist = false
		} else if line == "<BREAK>" {
			// Handles switching to a new song
			onTitle = true
			onArtist = false
			songChan <- currSong
			songs = append(songs, currSong)
			currID++
			currSong = song{id: currID}
		} else {
			// Append each word in the line to the current songs word slice
			currSong.words = append(currSong.words, strings.Split(line, " ")...)
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	// Force the input loop to finish in go processSong
	close(songChan)
	finish := time.Now()
	table := <-tableChan

	for key, value := range table {
		sort.Sort(value)
		if len(value) > 10 {
			value = value[0:10]
		}
		table[key] = value
	}

	console := bufio.NewScanner(os.Stdin)
	fmt.Println("Time passed: ", finish.Sub(begin))
	fmt.Println("Please enter your search term below:")
	for console.Scan() {
		line := console.Text()
		if strings.Contains(line, " ") {
			fmt.Println("Please input a valid single word")
		} else if line == "<BREAK>" {
			return
		} else {
			line = strings.ToLower(line)
			retrieveLyric(line, table, songs)
		}
	}
}

func processSong(in chan song, out chan map[string]words) {
	table := make(map[string]words)

	re := regexp.MustCompile("[\\W_]")
	// Iterate through songs that are sent into the channel
	for song := range in {
		// Iterate over the words in each song
		for i, w := range song.words {
			key := re.ReplaceAllString(w, "")
			key = strings.ToLower(key)
			node, ok := table[key]
			// If the node was present in the table
			if ok {
				// Check to see if we already have an occurence of this word
				if node[len(node)-1].songID != song.id {
					currWord := word{songID: song.id, word: w}
					currWord.positions = append(currWord.positions, i)
					sort.Sort(node)
					if len(node) > 10 {
						node = node[0:10]
					}
					node = append(node, currWord)
					table[key] = node
				} else {
					currWord := node[len(node)-1]
					currWord.positions = append(currWord.positions, i)
					node[len(node)-1] = currWord
					table[key] = node
				}
			} else {
				// Make a new word node if it wasn't present in the table
				node := make(words, 0)
				currWord := word{songID: song.id, word: w}
				currWord.positions = append(currWord.positions, i)
				node = append(node, currWord)
				table[key] = node
			}
		}
	}
	out <- table
}

func retrieveLyric(lyric string, table map[string]words, songs []song) {
	occurences, ok := table[lyric]
	if ok {
		for _, node := range occurences {
			song := songs[node.songID]
			for _, position := range node.positions {
				fmt.Println("Title:", song.title)
				fmt.Println("Arist:", song.artist)
				var start int
				var end int
				if position < 5 {
					start = 0
				} else {
					start = position - 5
				}
				if position+6 >= len(song.words) {
					end = len(song.words)
				} else {
					end = position + 6
				}
				fmt.Println("Context:", strings.Join(song.words[start:end], " "))
			}
		}
		fmt.Println("<END-OF-REPORT>")
	} else {
		fmt.Println("Search term not found")
	}
}
