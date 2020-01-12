// Package telnetapi provides a telnet server connection handler that gets
// called when a new telnet session is initiated.  It creates/managers
// individual telnet connections and parses/forwards telnet commands to
// those connections.
package telnetapi

import (
	"bytes"
	"chatserver/model"
	"chatserver/model/subs"
	"chatserver/telnetconn"
	"log"
	"strconv"
	"strings"
	"sync"

	oi "github.com/reiver/go-oi"
	gotelnet "github.com/reiver/go-telnet"
)

// ConnectionHandler holds data that needs to be forwarded/used for the
// individual telnet connections
type ConnectionHandler struct {
	model      *model.Model
	subsEngine *subs.Engine
}

// NewConnectionHandler creates/initializes/returns a new ConnectionHandler
func NewConnectionHandler(model *model.Model, subsEngine *subs.Engine) *ConnectionHandler {
	handler := ConnectionHandler{
		model:      model,
		subsEngine: subsEngine,
	}

	return &handler
}

// ServeTELNET satisfies the go-telnet Handler interface and is called
// whenever a new telnet session is initiated.  It will create a new telnet
// connection and parse/forward telnet commands to that connection.
func (h *ConnectionHandler) ServeTELNET(ctx gotelnet.Context, writer gotelnet.Writer, reader gotelnet.Reader) {
	connChan := make(chan error)

	// We need a mutex for each connection in case we get printLinesCallback called from multiple goroutines
	var connMutex sync.Mutex
	printLinesCallback := func(lines []string) {
		connMutex.Lock()
		defer connMutex.Unlock()

		// Write the new text to the telnet client
		for _, line := range lines {
			_, err := oi.LongWriteString(writer, line+"\r\n")
			if err != nil {
				connChan <- err
				return
			}
		}
	}

	// Create a new telnet connection
	telnetConn := telnetconn.NewTelnetConn(h.model, printLinesCallback)

	// Connect it to the subscription engine
	err := h.subsEngine.Connect(telnetConn)
	if err != nil {
		log.Fatal(err)
	}

	// Handle the new connection
	go h.handleConn(ctx, writer, reader, telnetConn, connChan)

	// Wait for the handler to exit
	err = <-connChan
	if err != nil {
		log.Fatal(err)
	}

	// Clean up the subscriptions
	err = h.subsEngine.Disconnect(telnetConn)
	if err != nil {
		log.Fatal(err)
	}
}

func (h *ConnectionHandler) writePrompt(writer gotelnet.Writer) error {
	var prompt bytes.Buffer
	prompt.WriteString("$ ")
	promptBytes := prompt.Bytes()

	// Print the prompt to the client
	_, err := oi.LongWrite(writer, promptBytes)
	if err != nil {
		return err
	}

	return nil
}

func (h *ConnectionHandler) parseHelpCmd(telnetConn *telnetconn.TelnetConn, writer gotelnet.Writer, fields []string) error {
	if _, err := oi.LongWriteString(writer, "'chatserver' commands:\r\n"); err != nil {
		return err
	}
	if _, err := oi.LongWriteString(writer, "\r\n"); err != nil {
		return err
	}
	if _, err := oi.LongWriteString(writer, "<message> - post a <message>\r\n"); err != nil {
		return err
	}
	if _, err := oi.LongWriteString(writer, "\r\n"); err != nil {
		return err
	}
	if _, err := oi.LongWriteString(writer, "/users - display users\r\n"); err != nil {
		return err
	}
	if _, err := oi.LongWriteString(writer, "/user <user> - change current user to <user>\r\n"); err != nil {
		return err
	}
	if _, err := oi.LongWriteString(writer, "/userinfo - display info about the current user\r\n"); err != nil {
		return err
	}
	if _, err := oi.LongWriteString(writer, "/createuser <user> - create a new <user>\r\n"); err != nil {
		return err
	}
	if _, err := oi.LongWriteString(writer, "/deleteuser <user> - delete an existing <user>\r\n"); err != nil {
		return err
	}
	if _, err := oi.LongWriteString(writer, "/blockuser <user> - block posts from <user>\r\n"); err != nil {
		return err
	}
	if _, err := oi.LongWriteString(writer, "/unblockuser <user> - unblock posts from <user>\r\n"); err != nil {
		return err
	}
	if _, err := oi.LongWriteString(writer, "/channels - display channels\r\n"); err != nil {
		return err
	}
	if _, err := oi.LongWriteString(writer, "/channel <channel> - change current channel to <channel>\r\n"); err != nil {
		return err
	}
	if _, err := oi.LongWriteString(writer, "/channelinfo - display info about the current channel\r\n"); err != nil {
		return err
	}
	if _, err := oi.LongWriteString(writer, "/channelhistory <num messages> - show <num messages> of current channel history (-1 for all)\r\n"); err != nil {
		return err
	}
	if _, err := oi.LongWriteString(writer, "/createchannel <channel> - create a new <channel>\r\n"); err != nil {
		return err
	}
	if _, err := oi.LongWriteString(writer, "/deletechannel <channel> - delete an existing <channel>\r\n"); err != nil {
		return err
	}
	if _, err := oi.LongWriteString(writer, "/exit - exit\r\n"); err != nil {
		return err
	}

	return nil
}

func (h *ConnectionHandler) parseUsersCmd(telnetConn *telnetconn.TelnetConn, writer gotelnet.Writer, fields []string) error {
	if len(fields) != 1 {
		if _, err := oi.LongWriteString(writer, "error: unknown /users option\r\n"); err != nil {
			return err
		}

		return nil
	}

	telnetConn.ShowUsers()
	return nil
}

func (h *ConnectionHandler) parseUserCmd(telnetConn *telnetconn.TelnetConn, writer gotelnet.Writer, fields []string) error {
	if len(fields) == 1 {
		if _, err := oi.LongWriteString(writer, "error: must provide a <user>\r\n"); err != nil {
			return err
		}

		return nil
	}

	if len(fields) > 2 {
		if _, err := oi.LongWriteString(writer, "error: <user> must not contain spaces\r\n"); err != nil {
			return err
		}

		return nil
	}

	telnetConn.SwitchUser(fields[1])
	return nil
}

func (h *ConnectionHandler) parseUserInfoCmd(telnetConn *telnetconn.TelnetConn, writer gotelnet.Writer, fields []string) error {
	if len(fields) != 1 {
		if _, err := oi.LongWriteString(writer, "error: unknown /userinfo option\r\n"); err != nil {
			return err
		}

		return nil
	}

	telnetConn.ShowUserInfo()
	return nil
}

func (h *ConnectionHandler) parseCreateUserCmd(telnetConn *telnetconn.TelnetConn, writer gotelnet.Writer, fields []string) error {
	if len(fields) == 1 {
		if _, err := oi.LongWriteString(writer, "error: must provide a <user>\r\n"); err != nil {
			return err
		}

		return nil
	}

	if len(fields) > 2 {
		if _, err := oi.LongWriteString(writer, "error: <user> must not contain spaces\r\n"); err != nil {
			return err
		}

		return nil
	}

	telnetConn.CreateUser(fields[1])
	return nil
}

func (h *ConnectionHandler) parseDeleteUserCmd(telnetConn *telnetconn.TelnetConn, writer gotelnet.Writer, fields []string) error {
	if len(fields) == 1 {
		if _, err := oi.LongWriteString(writer, "error: must provide a <user>\r\n"); err != nil {
			return err
		}

		return nil
	}

	if len(fields) > 2 {
		if _, err := oi.LongWriteString(writer, "error: <user> must not contain spaces\r\n"); err != nil {
			return err
		}

		return nil
	}

	telnetConn.DeleteUser(fields[1])
	return nil
}

func (h *ConnectionHandler) parseBlockUserCmd(telnetConn *telnetconn.TelnetConn, writer gotelnet.Writer, fields []string) error {
	if len(fields) == 1 {
		if _, err := oi.LongWriteString(writer, "error: must provide a <user>\r\n"); err != nil {
			return err
		}

		return nil
	}

	if len(fields) > 2 {
		if _, err := oi.LongWriteString(writer, "error: <user> must not contain spaces\r\n"); err != nil {
			return err
		}

		return nil
	}

	telnetConn.BlockUser(fields[1])
	return nil
}

func (h *ConnectionHandler) parseUnblockUserCmd(telnetConn *telnetconn.TelnetConn, writer gotelnet.Writer, fields []string) error {
	if len(fields) == 1 {
		if _, err := oi.LongWriteString(writer, "error: must provide a <user>\r\n"); err != nil {
			return err
		}

		return nil
	}

	if len(fields) > 2 {
		if _, err := oi.LongWriteString(writer, "error: <user> must not contain spaces\r\n"); err != nil {
			return err
		}

		return nil
	}

	telnetConn.UnblockUser(fields[1])
	return nil
}

func (h *ConnectionHandler) parseChannelsCmd(telnetConn *telnetconn.TelnetConn, writer gotelnet.Writer, fields []string) error {
	if len(fields) != 1 {
		if _, err := oi.LongWriteString(writer, "error: unknown /channels option\r\n"); err != nil {
			return err
		}

		return nil
	}

	telnetConn.ShowChannels()
	return nil
}

func (h *ConnectionHandler) parseChannelCmd(telnetConn *telnetconn.TelnetConn, writer gotelnet.Writer, fields []string) error {
	if len(fields) == 1 {
		if _, err := oi.LongWriteString(writer, "error: must provide a <channel>\r\n"); err != nil {
			return err
		}

		return nil
	}

	if len(fields) > 2 {
		if _, err := oi.LongWriteString(writer, "error: <channel> must not contain spaces\r\n"); err != nil {
			return err
		}

		return nil
	}

	telnetConn.SwitchChannel(fields[1])
	return nil
}

func (h *ConnectionHandler) parseChannelInfoCmd(telnetConn *telnetconn.TelnetConn, writer gotelnet.Writer, fields []string) error {
	if len(fields) != 1 {
		if _, err := oi.LongWriteString(writer, "error: unknown /channelinfo option\r\n"); err != nil {
			return err
		}

		return nil
	}

	telnetConn.ShowChannelInfo()
	return nil
}

func (h *ConnectionHandler) parseChannelHistoryCmd(telnetConn *telnetconn.TelnetConn, writer gotelnet.Writer, fields []string) error {
	if len(fields) == 1 {
		if _, err := oi.LongWriteString(writer, "error: must provide <num messages>\r\n"); err != nil {
			return err
		}

		return nil
	}

	if len(fields) > 2 {
		if _, err := oi.LongWriteString(writer, "error: unknown /channelhistory option\r\n"); err != nil {
			return err
		}

		return nil
	}

	numMessages, err := strconv.Atoi(fields[1])
	if err != nil || numMessages < -1 {
		if _, err := oi.LongWriteString(writer, "error: invalid <num messages>\r\n"); err != nil {
			return err
		}

		return nil
	}

	telnetConn.ShowChannelHistory(numMessages)
	return nil
}

func (h *ConnectionHandler) parseCreateChannelCmd(telnetConn *telnetconn.TelnetConn, writer gotelnet.Writer, fields []string) error {
	if len(fields) == 1 {
		if _, err := oi.LongWriteString(writer, "error: must provide a <channel>\r\n"); err != nil {
			return err
		}

		return nil
	}

	if len(fields) > 2 {
		if _, err := oi.LongWriteString(writer, "error: <channel> must not contain spaces\r\n"); err != nil {
			return err
		}

		return nil
	}

	telnetConn.CreateChannel(fields[1])
	return nil
}

func (h *ConnectionHandler) parseDeleteChannelCmd(telnetConn *telnetconn.TelnetConn, writer gotelnet.Writer, fields []string) error {
	if len(fields) == 1 {
		if _, err := oi.LongWriteString(writer, "error: must provide a <channel>\r\n"); err != nil {
			return err
		}

		return nil
	}

	if len(fields) > 2 {
		if _, err := oi.LongWriteString(writer, "error: <channel> must not contain spaces\r\n"); err != nil {
			return err
		}

		return nil
	}

	telnetConn.DeleteChannel(fields[1])
	return nil
}

func (h *ConnectionHandler) handleConn(ctx gotelnet.Context, writer gotelnet.Writer, reader gotelnet.Reader, telnetConn *telnetconn.TelnetConn, c chan error) {
	// NOTE: Assume all write errors mean the session has ended and should be swallowed
	err := h.writePrompt(writer)
	if err != nil {
		c <- nil
		return
	}

	// Create the buffer to hold user input
	var buffer [1]byte
	p := buffer[:]
	var line bytes.Buffer

	for {
		// Read 1 byte.
		n, err := reader.Read(p)
		if err != nil {
			c <- nil
			return
		}

		if n <= 0 {
			continue
		}

		line.WriteByte(p[0])

		// Newline specifies the end of a sent message.  Parse it.
		if '\n' == p[0] {
			lineString := line.String()

			fields := strings.Fields(lineString)
			if len(fields) > 0 && lineString != "\r\n" {
				// Parse the message
				command := fields[0]

				err = nil
				switch command {
				case "/help":
					err = h.parseHelpCmd(telnetConn, writer, fields)
				case "/users":
					err = h.parseUsersCmd(telnetConn, writer, fields)
				case "/user":
					err = h.parseUserCmd(telnetConn, writer, fields)
				case "/userinfo":
					err = h.parseUserInfoCmd(telnetConn, writer, fields)
				case "/createuser":
					err = h.parseCreateUserCmd(telnetConn, writer, fields)
				case "/deleteuser":
					err = h.parseDeleteUserCmd(telnetConn, writer, fields)
				case "/blockuser":
					err = h.parseBlockUserCmd(telnetConn, writer, fields)
				case "/unblockuser":
					err = h.parseUnblockUserCmd(telnetConn, writer, fields)
				case "/channels":
					err = h.parseChannelsCmd(telnetConn, writer, fields)
				case "/channel":
					err = h.parseChannelCmd(telnetConn, writer, fields)
				case "/channelinfo":
					err = h.parseChannelInfoCmd(telnetConn, writer, fields)
				case "/channelhistory":
					err = h.parseChannelHistoryCmd(telnetConn, writer, fields)
				case "/createchannel":
					err = h.parseCreateChannelCmd(telnetConn, writer, fields)
				case "/deletechannel":
					err = h.parseDeleteChannelCmd(telnetConn, writer, fields)
				case "/exit":
					c <- nil
					return
				default:
					if command[0] == '/' {
						_, err = oi.LongWriteString(writer, "error: unknown command\r\n")
					} else {
						telnetConn.PostMessage(strings.TrimSuffix(lineString, "\r\n"))
					}
				}

				if err != nil {
					c <- nil
					return
				}
			}

			// Print the prompt
			line.Reset()
			err = h.writePrompt(writer)
			if err != nil {
				c <- nil
				return
			}
		}
	}
}
