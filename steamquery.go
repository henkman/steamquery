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
type Response struct {
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

func Query(address *net.UDPAddr) (Response, error) {
	c, err := net.DialUDP("udp", nil, address)
	if err != nil {
		return Response{}, err
	}
	c.Write([]byte("\xFF\xFF\xFF\xFFTSource Engine Query\x00"))
	var buf [2 * 1024]byte
	n, _ := c.Read(buf[:])
	c.Close()
	if n <= 0 {
		return Response{}, errors.New("got invalid response")
	}
	var r Response
	o := 5 // skip "\xFF\xFF\xFF\xFFI"
	r.Protocol = buf[o]
	o++
	nb := bytes.IndexByte(buf[o:], 0)
	if nb == -1 {
		return Response{}, errors.New("got invalid response")
	}
	r.Name = string(buf[o : o+nb])
	o += nb + 1
	nb = bytes.IndexByte(buf[o:], 0)
	if nb == -1 {
		return Response{}, errors.New("got invalid response")
	}
	r.Map = string(buf[o : o+nb])
	o += nb + 1
	nb = bytes.IndexByte(buf[o:], 0)
	if nb == -1 {
		return Response{}, errors.New("got invalid response")
	}
	r.Folder = string(buf[o : o+nb])
	o += nb + 1
	nb = bytes.IndexByte(buf[o:], 0)
	if nb == -1 {
		return Response{}, errors.New("got invalid response")
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
		return Response{}, errors.New("got invalid response")
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
			return Response{}, errors.New("got invalid response")
		}
		r.SourceTV.Name = string(buf[o : o+nb])
		o += nb + 1
	}
	if extra&0x20 != 0 {
		nb = bytes.IndexByte(buf[o:], 0)
		if nb == -1 {
			return Response{}, errors.New("got invalid response")
		}
		r.Keywords = string(buf[o : o+nb])
		o += nb + 1
	}
	if extra&0x01 != 0 {
		r.GameID = binary.LittleEndian.Uint64(buf[o:])
		o += 8
	}
	return r, nil
}

func QueryString(address string) (Response, error) {
	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return Response{}, err
	}
	return Query(addr)
}

func QueryPlayers(address *net.UDPAddr) ([]Player, error) {
	c, err := net.DialUDP("udp", nil, address)
	if err != nil {
		return nil, err
	}
	c.Write([]byte("\xFF\xFF\xFF\xFFU\xFF\xFF\xFF\xFF"))
	var buf [2 * 1024]byte
	n, _ := c.Read(buf[:])
	if n < 9 {
		return nil, errors.New("got invalid response")
	}
	buf[4] = 'U'
	c.Write(buf[:n])
	n, _ = c.Read(buf[:])
	if n < 2 {
		return nil, errors.New("got invalid response")
	}
	c.Close()
	o := 5 // skip "\xFF\xFF\xFF\xFFD"
	np := buf[o]
	o++
	players := make([]Player, 0, np)
	for i := byte(0); i < np; i++ {
		var p Player
		o++ // skip index
		nb := bytes.IndexByte(buf[o:], 0)
		if nb == -1 {
			return nil, errors.New("got invalid response")
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
	return players, nil
}

func QueryPlayersString(address string) ([]Player, error) {
	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return nil, err
	}
	return QueryPlayers(addr)
}
