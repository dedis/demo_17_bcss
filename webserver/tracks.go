package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"sync"

	"gopkg.in/dedis/onet.v1/log"
)

type Track struct {
	ID      int
	Title   string
	Persons string
	Date    string
}

// custom database yay
type databaseStruct struct {
	DB map[int]*entryStruct
	sync.Mutex
}

// Represents a row in the database including the votes for the track
type entryStruct struct {
	Track
	// map of tag => vote status
	Votes []voteStruct
}

type voteStruct struct {
	Tag []byte
	// True = voted for , False = voted against
	Vote bool
}

// this struct represents the entries that are given to the javascript
type entryJSON struct {
	Track
	Voted bool
	Up    int
	Down  int
}

// this struct is just a wrapper to easily sort a list of entryJSON
type entriesJSON []entryJSON

func (e *entriesJSON) Len() int {
	return len(*e)
}

func (e *entriesJSON) Less(i, j int) bool {
	a := (*e)[i]
	b := (*e)[j]
	return a.ID < b.ID
}

func (e *entriesJSON) Swap(i, j int) {
	tmp := (*e)[i]
	(*e)[i] = (*e)[j]
	(*e)[j] = tmp
}

func newDatabase() *databaseStruct {
	return &databaseStruct{DB: map[int]*entryStruct{}}
}

// Returns the JSON representation in "entryJSON" format
// with information including whether this tag has voted or not
// The result is usually sent to a javascript engine.
func (d *databaseStruct) JSON(tag []byte) ([]byte, error) {
	d.Lock()
	defer d.Unlock()

	var eJSONs entriesJSON
	for _, entry := range d.DB {
		var voted bool
		var up, down = 0, 0
		// count the votes
		for _, v := range entry.Votes {
			if bytes.Equal(tag, v.Tag) {
				voted = v.Vote
			}
			if v.Vote {
				up++
			} else {
				down++
			}
		}
		eJSON := entryJSON{
			Track: entry.Track,
			Up:    up,
			Down:  down,
			Voted: voted,
		}
		eJSONs = append(eJSONs, eJSON)
	}
	sort.Stable(&eJSONs)
	return json.Marshal(eJSONs)
}

// Create a new vote entry or update an already existing vote (if different) or
// return an error
func (d *databaseStruct) VoteOrError(id int, tag []byte, vote bool) error {
	d.Lock()
	defer d.Unlock()
	e, ok := d.DB[id]
	if !ok {
		return errors.New("invalid entry id")
	}
	// iterate over all the votes for this entry
	for i, t := range e.Votes {
		if bytes.Equal(tag, t.Tag) {
			if vote == t.Vote {
				return errors.New("users already voted")
			}
			e.Votes[i].Vote = vote
			return nil
		}
	}
	e.Votes = append(e.Votes, voteStruct{tag, vote})
	fmt.Println("Voted for ", e.Title, " from ", hex.EncodeToString(tag))
	return nil
}

type Tracks []entryStruct

// load the tracks with the associated votes if any.
func (d *databaseStruct) Load(fileName string) {
	d.Lock()
	defer d.Unlock()
	file, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}

	var tracks Tracks
	if err := json.NewDecoder(file).Decode(&tracks); err != nil {
		panic(err)
	}

	for _, track := range tracks {
		t := Track{
			ID:      track.ID,
			Persons: track.Persons,
			Title:   track.Title,
			Date:    track.Date}

		d.DB[track.Track.ID] = &entryStruct{Track: t, Votes: track.Votes}
	}
	log.Info("[+] Loaded ", len(d.DB), " tracks")
}

// Save stores the votes for later usage
func (d *databaseStruct) Save(fullName string) error {
	file, err := os.OpenFile(fullName, os.O_RDWR+os.O_CREATE, 0660)
	if err != nil {
		return err
	}
	var tracks Tracks
	for _, track := range d.DB {
		tracks = append(tracks, *track)
	}
	if err = json.NewEncoder(file).Encode(tracks); err != nil {
		return err
	}
	return file.Close()
}

func (t *Track) String() string {
	return fmt.Sprintf("%d: %s %s in %s", t.ID, t.Title, t.Persons, t.Date)
}

func (e *entryStruct) String() string {
	var up, down int
	for _, v := range e.Votes {
		if v.Vote {
			up += 1
		} else {
			down += 1
		}
	}
	return fmt.Sprintf("%s\n\t%d up, %d down", e.Track.String(), up, down)
}

func (d *databaseStruct) String() string {
	var b bytes.Buffer
	for _, e := range d.DB {
		fmt.Println(e.Track.String())
		//fmt.Fprintf(&b, e.String()+"\n")
	}
	return b.String()
}
