/*
* This is a template for creating an app. It only has one command which
* prints out the name of the app.
 */
package main

import (
	"os"

	"io/ioutil"

	"path"

	"errors"

	"encoding/base64"

	"bytes"

	"net"

	"fmt"

	"strings"

	"github.com/BurntSushi/toml"
	"github.com/dedis/demo_17_bcss/pop/service"
	"gopkg.in/dedis/cothority.v1/cosi/check"
	"gopkg.in/dedis/crypto.v0/abstract"
	"gopkg.in/dedis/crypto.v0/anon"
	"gopkg.in/dedis/crypto.v0/random"
	"gopkg.in/dedis/onet.v1/app"
	"gopkg.in/dedis/onet.v1/log"
	"gopkg.in/dedis/onet.v1/network"
	"gopkg.in/urfave/cli.v1"
)

func init() {
	network.RegisterMessage(Config{})
}

var client *service.Client

// Config represents either a manager or an attendee configuration.
type Config struct {
	Private abstract.Scalar
	Public  abstract.Point
	Index   int
	Address network.Address
	Final   *service.FinalStatement
}

var mainConfig *Config
var fileConfig string

func main() {
	appCli := cli.NewApp()
	appCli.Name = "SSH keystore client"
	appCli.Usage = "Connects to a ssh-keystore-server and updates/changes information"
	appCli.Version = "0.3"
	appCli.Commands = []cli.Command{
		commandOrg,
		commandClient,
		{
			Name:      "check",
			Aliases:   []string{"c"},
			Usage:     "Check if the servers in the group definition are up and running",
			ArgsUsage: "group.toml",
			Action:    checkConfig,
		},
	}
	appCli.Flags = []cli.Flag{
		cli.IntFlag{
			Name:  "debug,d",
			Value: 0,
			Usage: "debug-level: 1 for terse, 5 for maximal",
		},
		cli.StringFlag{
			Name:  "config,c",
			Value: "~/.config/cothority/pop",
			Usage: "The configuration-directory of pop",
		},
	}
	appCli.Before = func(c *cli.Context) error {
		log.SetDebugVisible(c.Int("debug"))
		client = service.NewClient()
		fileConfig = path.Join(c.String("config"), "config.bin")
		readConfig()
		return nil
	}
	appCli.Run(os.Args)
}

// links this pop to a cothority
func orgLink(c *cli.Context) error {
	log.Info("Org: Link")
	if c.NArg() == 0 {
		log.Fatal("Please give an IP and optionally a pin")
	}
	newConfig()
	host, port, err := net.SplitHostPort(c.Args().First())
	if err != nil {
		return err
	}
	addrs, err := net.LookupHost(host)
	if err != nil {
		return err
	}
	addr := network.NewAddress(network.PlainTCP, fmt.Sprintf("%s:%s", addrs[0], port))
	pin := c.Args().Get(1)
	if err := client.Pin(addr, pin, mainConfig.Public); err != nil {
		if err.ErrorCode() == service.ErrorWrongPIN && pin == "" {
			log.Info("Please read PIN in server-log")
			return nil
		}
		return err
	}
	mainConfig.Address = addr
	log.Info("Successfully linked with", addr)
	writeConfig()
	return nil
}

// sets up a configuration
func orgConfig(c *cli.Context) error {
	log.Info("Org: Config", mainConfig.Address.String())
	if c.NArg() != 2 {
		log.Fatal("Please give pop_desc.toml and group.toml")
	}
	if mainConfig.Address.String() == "" {
		log.Fatal("No address")
		return errors.New("No address found - please link first")
	}
	desc := &service.PopDesc{}
	pdFile := c.Args().First()
	buf, err := ioutil.ReadFile(pdFile)
	log.ErrFatal(err, "While reading", pdFile)
	_, err = toml.Decode(string(buf), desc)
	log.ErrFatal(err, "While decoding", pdFile)
	group := readGroup(c.Args().Get(1))
	desc.Roster = group.Roster
	log.Info("Hash of config is:", base64.StdEncoding.EncodeToString(desc.Hash()))
	//log.ErrFatal(check.Servers(group), "Couldn't check servers")
	log.ErrFatal(client.StoreConfig(mainConfig.Address, desc))
	mainConfig.Final.Desc = desc
	mainConfig.Final.Attendees = []abstract.Point{}
	mainConfig.Final.Signature = []byte{}
	writeConfig()
	return nil
}

// adds a public key to the list
func orgPublic(c *cli.Context) error {
	log.Info("Org: Adding public keys", c.Args().First())
	if c.NArg() < 1 {
		log.Fatal("Please give a public key")
	}
	str := c.Args().First()
	if !strings.HasPrefix(str, "[") {
		str = "[" + str + "]"
	}
	str = strings.Replace(str, "\"", "", -1)
	str = strings.Replace(str, "[", "", -1)
	str = strings.Replace(str, "]", "", -1)
	str = strings.Replace(str, "\\", "", -1)
	log.Print(str)
	keys := strings.Split(str, ",")
	for _, k := range keys {
		pub := service.B64ToPoint(k)
		if pub == nil {
			log.Fatal("Couldn't parse public key:", k)
		}
		for _, p := range mainConfig.Final.Attendees {
			if p.Equal(pub) {
				log.Fatal("This key already exists")
			}
		}
		mainConfig.Final.Attendees = append(mainConfig.Final.Attendees, pub)
	}
	writeConfig()
	return nil
}

// finalizes the statement
func orgFinal(c *cli.Context) error {
	log.Info("Org: Final")
	if len(mainConfig.Final.Attendees) == 0 {
		log.Fatal("No attendees stored - first store at least one")
	}
	if mainConfig.Address == "" {
		log.Fatal("Not linked")
	}
	if len(mainConfig.Final.Signature) > 0 {
		log.Info("Final statement already here:\n", "\n"+mainConfig.Final.ToToml())
		return nil
	}
	fs, err := client.Finalize(mainConfig.Address, mainConfig.Final.Desc, mainConfig.Final.Attendees)
	log.ErrFatal(err)
	mainConfig.Final = fs
	writeConfig()
	log.Info("Created final statement:\n", "\n"+mainConfig.Final.ToToml())
	return nil
}

// creates a new private/public pair
func clientCreate(c *cli.Context) error {
	priv := network.Suite.NewKey(random.Stream)
	pub := network.Suite.Point().Mul(nil, priv)
	log.Infof("Private: %s\nPublic: %s",
		service.ScalarToB64(priv),
		service.PointToB64(pub))
	return nil
}

// joins a poparty
func clientJoin(c *cli.Context) error {
	log.Info("Client: join")
	if c.NArg() < 2 {
		log.Fatal("Please give final.toml and private key.")
	}
	finalName := c.Args().First()
	privStr := c.Args().Get(1)
	privBuf, err := base64.StdEncoding.DecodeString(privStr)
	log.ErrFatal(err)
	priv := network.Suite.Scalar()
	log.ErrFatal(priv.UnmarshalBinary(privBuf))
	buf, err := ioutil.ReadFile(finalName)
	log.ErrFatal(err)
	mainConfig.Final = service.NewFinalStatementFromString(string(buf))
	if mainConfig.Final == nil {
		log.Fatal("Couldn't parse final statement")
	}
	mainConfig.Private = priv
	mainConfig.Public = network.Suite.Point().Mul(nil, priv)
	mainConfig.Index = -1
	for i, p := range mainConfig.Final.Attendees {
		if p.Equal(mainConfig.Public) {
			log.Info("Found public key at index", i)
			mainConfig.Index = i
		}
	}
	if mainConfig.Index == -1 {
		log.Fatal("Didn't find our public key in the final statement!")
	}
	writeConfig()
	log.Info("Stored new final statement and key.")
	return nil
}

// signs a message + context
func clientSign(c *cli.Context) error {
	log.Info("Client: sign")
	if mainConfig.Index == -1 {
		log.Fatal("No public key stored.")
	}
	if c.NArg() < 2 {
		log.Fatal("Please give msg and context")
	}
	msg := []byte(c.Args().First())
	ctx := []byte(c.Args().Get(1))

	Set := anon.Set(mainConfig.Final.Attendees)
	sigtag := anon.Sign(network.Suite, random.Stream, msg,
		Set, ctx, mainConfig.Index, mainConfig.Private)
	sig := sigtag[:len(sigtag)-32]
	tag := sigtag[len(sigtag)-32:]
	log.Infof("\nSignature: %s\nTag: %s", base64.StdEncoding.EncodeToString(sig),
		base64.StdEncoding.EncodeToString(tag))
	return nil
}

// verifies a signature and tag
func clientVerify(c *cli.Context) error {
	log.Info("Client: verify")
	if mainConfig.Index == -1 {
		log.Fatal("No public key stored")
	}
	if c.NArg() < 4 {
		log.Fatal("Please give a msg, context, signature and a tag")
	}
	msg := []byte(c.Args().First())
	ctx := []byte(c.Args().Get(1))
	sig, err := base64.StdEncoding.DecodeString(c.Args().Get(2))
	log.ErrFatal(err)
	tag, err := base64.StdEncoding.DecodeString(c.Args().Get(3))
	log.ErrFatal(err)
	sigtag := append(sig, tag...)
	ctag, err := anon.Verify(network.Suite, msg,
		anon.Set(mainConfig.Final.Attendees), ctx, sigtag)
	log.ErrFatal(err)
	if !bytes.Equal(tag, ctag) {
		log.Fatalf("Tag and calculated tag are not equal:\n%x - %x", tag, ctag)
	}
	log.Info("Successfully verified signature and tag")
	return nil
}

func readGroup(name string) *app.Group {
	f, err := os.Open(name)
	log.ErrFatal(err, "Couldn't open group definition file")
	group, err := app.ReadGroupDescToml(f)
	log.ErrFatal(err, "Error while reading group definition file", err)
	if len(group.Roster.List) == 0 {
		log.ErrFatalf(err, "Empty entity or invalid group defintion in: %s",
			name)
	}
	return group
}

// checkConfig contacts all servers and verifies if it receives a valid
// signature from each.
func checkConfig(c *cli.Context) error {
	return check.Config(c.Args().First(), false)
}

func newConfig() {
	mainConfig = &Config{
		Private: network.Suite.NewKey(random.Stream),
		Final: &service.FinalStatement{
			Attendees: []abstract.Point{},
			Signature: []byte{},
		},
		Index: -1,
	}
	mainConfig.Public = network.Suite.Point().Mul(nil, mainConfig.Private)
}

func readConfig() {
	file := app.TildeToHome(fileConfig)
	if _, err := os.Stat(file); err != nil {
		newConfig()
		return
	}
	buf, err := ioutil.ReadFile(file)
	if err == nil {
		_, msg, err := network.Unmarshal(buf)
		if err == nil {
			var ok bool
			mainConfig, ok = msg.(*Config)
			if ok {
				log.Lvlf2("Read config-file: %v", mainConfig)
				return
			}
		}
	}
	log.Fatal("Couldn't read", file, "- please remove it.")
}

func writeConfig() {
	buf, err := network.Marshal(mainConfig)
	log.ErrFatal(err)
	file := app.TildeToHome(fileConfig)
	os.MkdirAll(path.Dir(file), 0770)
	log.ErrFatal(ioutil.WriteFile(file, buf, 0660))
}
