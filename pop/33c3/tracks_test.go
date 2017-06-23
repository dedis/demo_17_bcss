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
	schedule33c3 = "schedule.json"
	prgid        = 7911
)

func TestDatabase__VotesSave(t *testing.T) {
	db := newDatabase()
	db.load(schedule33c3)

	db.DB[prgid].Votes = []voteStruct{}
	tmpfile, err := ioutil.TempFile("", "db")
	log.ErrFatal(err)
	tmpfile.Close()
	defer os.Remove(tmpfile.Name())
	log.ErrFatal(db.VotesSave(tmpfile.Name()))
}

func TestDatabase__VotesLoad(t *testing.T) {
	db := newDatabase()
	db.load(schedule33c3)
	db.DB[prgid].Votes = []voteStruct{{[]byte{}, true}}
	tmpfile, err := ioutil.TempFile("", "db")
	log.ErrFatal(err)
	tmpfile.Close()
	defer os.Remove(tmpfile.Name())
	log.ErrFatal(db.VotesSave(tmpfile.Name()))

	db2 := newDatabase()
	log.ErrFatal(db2.VotesLoad(schedule33c3, tmpfile.Name()))
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
