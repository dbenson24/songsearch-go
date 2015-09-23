# Derek Benson
## GO implementation of Song Search

This is an implementation of HW6 from Comp 15 data structures in the Go
programming language. I used 2 structs to hold all of the necessary information
for the program.

### song
- title
- artist
- a slice of words
- a reference to it's position in storage

### word
- word
- a reference to it's parent song
- a slice of positions in terms of where this word is located in it's songs lyrics
- implements the interface for Sort

## Execution
1. Cli app is initialized, readFile is called
2. readFile initializes a file stream and opens the input and output channels
for the processSong go routine, it then starts the processSong go routine.
3. readFile then scans through the input data, parsing each song. Once a song
is parsed, the completed song object gets pushed into the song channel where it
is processed by the processSong go routine.
4. readFile finishes moving through the input file and closes the input channel.
It then waits to receive output from the table channel.
5. After the finished table is received, the program goes through each entry
and makes sure that it is sorted properly.
6. Once the sorting is complete the input loop is initialized.
7. Upon receiving a valid word, the retrieveLyric function looks the word up
in the table, retrieve the array of songs, and prints the context of each occurence
of the word.
8. Input loop terminates after receiving the word "<BREAK"
