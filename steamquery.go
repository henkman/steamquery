package steamquery

import (
	"bytes"
	"encoding/binary"
	"errors"
	"math"
	"net"
	"time"
)

type ServerType byte

const (
	ServerType_Dedicated     ServerType = 'd'
	ServerType_NonDedicated  ServerType = 'l'
	ServerType_SourceTVRelay ServerType = 'p'
)

type Environment byte

const (
	Environment_Linux   Environment = 'l'
	Environment_Windows Environment = 'w'
	Environment_Mac     Environment = 'm' // or 'o'
)

// https://developer.valvesoftware.com/wiki/Source_Server_Queries#Response_Format
type Info struct {
	Protocol    byte   // Protocol version used by the server.
	Name        string // Name of the server.
	Map         string // Map the server has currently loaded.
	Folder      string // Name of the folder containing the game files.
	Game        string // Full name of the game.
	ID          uint16 // Steam Application ID of game.
	Players     byte   // Number of players on the server.
	MaxPlayers  byte   // Maximum number of players the server reports it can hold.
	Bots        byte   // Number of bots on the server.
	ServerType  ServerType
	Environment Environment
	Private     bool
	VAC         bool
	Version     string // Version of the game installed on the server.
	Port        uint16 // The server's game port number.
	SteamID     uint64 // Server's SteamID.
	SourceTV    struct {
		Port uint16 // Spectator port number
		Name string // Name of the spectator server
	}
	Keywords string // Tags that describe the game according to the server
	GameID   uint64 // The server's 64-bit GameID.
}

type Player struct {
	Name     string
	Score    int32
	Duration time.Duration
}

type Rule struct {
	Name  string
	Value string
}

func QueryInfo(address *net.UDPAddr) (Info, time.Duration, error) {
	c, err := net.DialUDP("udp", nil, address)
	if err != nil {
		return Info{}, time.Duration(0), err
	}
	var buf [8 * 1024]byte
	c.SetReadDeadline(time.Now().Add(time.Second * 5))
	before := time.Now()
	const REQ = "\xFF\xFF\xFF\xFFTSource Engine Query\x00"
	c.Write([]byte(REQ))
	n, err := c.Read(buf[:])
	ping := time.Since(before)
	if err != nil {
		return Info{}, time.Duration(0), err
	}
	if n == 9 && buf[4] == 'A' {
		var challenge [4]byte
		copy(challenge[:], buf[5:])
		c.Write(append([]byte(REQ), challenge[:]...))
		c.SetReadDeadline(time.Now().Add(time.Second * 5))
		n, err = c.Read(buf[:])
		if err != nil {
			c.Close()
			return Info{}, time.Duration(0), err
		}
	}
	c.Close()
	if n < 17 {
		return Info{}, time.Duration(0), errors.New("got invalid response")
	}
	var r Info
	o := 5 // skip "\xFF\xFF\xFF\xFFI"
	r.Protocol = buf[o]
	o++
	nb := bytes.IndexByte(buf[o:], 0)
	if nb == -1 {
		return Info{}, time.Duration(0), errors.New("got invalid response")
	}
	r.Name = string(buf[o : o+nb])
	o += nb + 1
	nb = bytes.IndexByte(buf[o:], 0)
	if nb == -1 {
		return Info{}, time.Duration(0), errors.New("got invalid response")
	}
	r.Map = string(buf[o : o+nb])
	o += nb + 1
	nb = bytes.IndexByte(buf[o:], 0)
	if nb == -1 {
		return Info{}, time.Duration(0), errors.New("got invalid response")
	}
	r.Folder = string(buf[o : o+nb])
	o += nb + 1
	nb = bytes.IndexByte(buf[o:], 0)
	if nb == -1 {
		return Info{}, time.Duration(0), errors.New("got invalid response")
	}
	r.Game = string(buf[o : o+nb])
	o += nb + 1
	r.ID = binary.LittleEndian.Uint16(buf[o:])
	o += 2
	r.Players = buf[o]
	o++
	r.MaxPlayers = buf[o]
	o++
	r.Bots = buf[o]
	o++
	r.ServerType = ServerType(buf[o])
	o++
	if buf[o] == 'o' {
		r.Environment = Environment_Mac
	} else {
		r.Environment = Environment(buf[o])
	}
	o++
	r.Private = buf[o] == 1
	o++
	r.VAC = buf[o] == 1
	o++
	// EXTRA STUFF FOR "The Ship" not handled here
	nb = bytes.IndexByte(buf[o:], 0)
	if nb == -1 {
		return Info{}, time.Duration(0), errors.New("got invalid response")
	}
	r.Version = string(buf[o : o+nb])
	o += nb + 1
	extra := buf[o]
	o++
	if extra&0x80 != 0 {
		r.Port = binary.LittleEndian.Uint16(buf[o:])
		o += 2
	}
	if extra&0x10 != 0 {
		r.SteamID = binary.LittleEndian.Uint64(buf[o:])
		o += 8
	}
	if extra&0x40 != 0 {
		r.SourceTV.Port = binary.LittleEndian.Uint16(buf[o:])
		o += 2
		nb = bytes.IndexByte(buf[o:], 0)
		if nb == -1 {
			return Info{}, time.Duration(0), errors.New("got invalid response")
		}
		r.SourceTV.Name = string(buf[o : o+nb])
		o += nb + 1
	}
	if extra&0x20 != 0 {
		nb = bytes.IndexByte(buf[o:], 0)
		if nb == -1 {
			return Info{}, time.Duration(0), errors.New("got invalid response")
		}
		r.Keywords = string(buf[o : o+nb])
		o += nb + 1
	}
	if extra&0x01 != 0 {
		r.GameID = binary.LittleEndian.Uint64(buf[o:])
		o += 8
	}
	return r, ping, nil
}

func QueryInfoString(address string) (Info, time.Duration, error) {
	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return Info{}, time.Duration(0), err
	}
	return QueryInfo(addr)
}

func QueryPlayers(address *net.UDPAddr) ([]Player, time.Duration, error) {
	c, err := net.DialUDP("udp", nil, address)
	if err != nil {
		return nil, time.Duration(0), err
	}
	c.SetReadDeadline(time.Now().Add(time.Second * 5))
	var buf [2 * 1024]byte
	before := time.Now()
	c.Write([]byte("\xFF\xFF\xFF\xFFU\xFF\xFF\xFF\xFF"))
	n, err := c.Read(buf[:])
	ping := time.Since(before)
	if err != nil {
		c.Close()
		return nil, time.Duration(0), err
	}
	if n == 9 && buf[4] == 'A' {
		buf[4] = 'U'
		c.Write(buf[:n])
		c.SetReadDeadline(time.Now().Add(time.Second * 5))
		n, err = c.Read(buf[:])
		if err != nil {
			c.Close()
			return nil, time.Duration(0), err
		}
	}
	c.Close()
	if n < len("\xFF\xFF\xFF\xFFD0") {
		return nil, time.Duration(0), errors.New("got invalid response")
	}
	o := 5 // skip "\xFF\xFF\xFF\xFFD"
	np := buf[o]
	o++
	players := make([]Player, 0, np)
	for i := byte(0); i < np; i++ {
		var p Player
		o++ // skip index
		nb := bytes.IndexByte(buf[o:], 0)
		if nb == -1 {
			return nil, time.Duration(0), errors.New("got invalid response")
		}
		p.Name = string(buf[o : o+nb])
		o += nb + 1
		p.Score = int32(binary.LittleEndian.Uint32(buf[o:]))
		o += 4
		dur := math.Float32frombits(binary.LittleEndian.Uint32(buf[o:]))
		p.Duration = time.Duration(dur * 1000 * 1000 * 1000)
		o += 4
		if len(p.Name) != 0 {
			players = append(players, p)
		}
	}
	return players, ping, nil
}

func QueryPlayersString(address string) ([]Player, time.Duration, error) {
	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return nil, time.Duration(0), err
	}
	return QueryPlayers(addr)
}

func QueryRules(address *net.UDPAddr) ([]Rule, time.Duration, error) {
	c, err := net.DialUDP("udp", nil, address)
	if err != nil {
		return nil, time.Duration(0), err
	}
	var buf [8 * 1024]byte
	c.SetReadDeadline(time.Now().Add(time.Second * 5))
	before := time.Now()
	c.Write([]byte("\xFF\xFF\xFF\xFFV\xFF\xFF\xFF\xFF"))
	n, err := c.Read(buf[:])
	ping := time.Since(before)
	if err != nil {
		c.Close()
		return nil, time.Duration(0), err
	}
	if n == 9 && buf[4] == 'A' {
		buf[4] = 'V'
		c.Write(buf[:n])
		c.SetReadDeadline(time.Now().Add(time.Second * 5))
		n, err = c.Read(buf[:])
		if err != nil {
			c.Close()
			return nil, time.Duration(0), err
		}
	}
	c.Close()
	if n < len("\xFF\xFF\xFF\xFFE00") {
		return nil, time.Duration(0), errors.New("got invalid response")
	}
	o := 5 // skip "\xFF\xFF\xFF\xFFE"
	nr := binary.LittleEndian.Uint16(buf[o:])
	o += 2
	rules := make([]Rule, 0, nr)
	for i := uint16(0); i < nr; i++ {
		var r Rule
		nb := bytes.IndexByte(buf[o:], 0)
		if nb == -1 {
			return nil, time.Duration(0), errors.New("got invalid response")
		}
		r.Name = string(buf[o : o+nb])
		o += nb + 1
		nb = bytes.IndexByte(buf[o:], 0)
		if nb == -1 {
			return nil, time.Duration(0), errors.New("got invalid response")
		}
		r.Value = string(buf[o : o+nb])
		o += nb + 1
		if len(r.Name) != 0 {
			rules = append(rules, r)
		}
	}
	return rules, ping, nil
}

func QueryRulesString(address string) ([]Rule, time.Duration, error) {
	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return nil, time.Duration(0), err
	}
	return QueryRules(addr)
}
