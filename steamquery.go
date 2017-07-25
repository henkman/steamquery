package steamquery

import (
	"bytes"
	"encoding/binary"
	"errors"
	"net"
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
