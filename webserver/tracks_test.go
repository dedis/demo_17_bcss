package main

import (
	"io/ioutil"
	"testing"

	"os"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/dedis/onet.v1/log"
)

const (
	schedule = "tracks.json"
	prgid    = 8
)

func TestDatabase__VotesSave(t *testing.T) {
	db := newDatabase()
	db.Load(schedule)

	tmpfile, err := ioutil.TempFile("", "db")
	log.ErrFatal(err)
	tmpfile.Close()
	defer os.Remove(tmpfile.Name())
	log.ErrFatal(db.Save(tmpfile.Name()))
}

func TestDatabase__VotesLoad(t *testing.T) {
	db := newDatabase()
	db.Load(schedule)
	votes := []voteStruct{{[]byte{}, true}}
	db.DB[prgid].Votes = votes
	tmpfile, err := ioutil.TempFile("", "db")
	log.ErrFatal(err)
	tmpfile.Close()
	defer os.Remove(tmpfile.Name())
	log.ErrFatal(db.Save(tmpfile.Name()))

	db2 := newDatabase()
	db2.Load(tmpfile.Name())
	require.True(t, db2.DB[prgid].Votes[0].Vote)
}

func TestSessionStore__Save(t *testing.T) {
	st := newSessionStore()
	st.Sessions = [][]byte{}
	st.Nonces = [][]byte{}

	tmpfile, err := ioutil.TempFile("", "st")
	log.ErrFatal(err)
	tmpfile.Close()
	defer os.Remove(tmpfile.Name())
	st.Save(tmpfile.Name())

	st2 := newSessionStore()
	st2.Load(tmpfile.Name())
	assert.Equal(t, st.Sessions, st2.Sessions)
	assert.Equal(t, st.Nonces, st2.Nonces)
}
