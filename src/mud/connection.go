package mud

import ("net"
	"strings")

type UserConnection struct {
	socket net.Conn
	outOfBand bool
	done chan bool
	FromUser chan string
	ToUser chan string
	OnDisconnect func()
	State ConnectionState
	Data map[string]interface{}
}

/* 
 ConnectionState should be derived for state machine-type login processes.

 It operates on the "out-of-band" console, used for logins or any connection
 which isn't yet handled by a Player ExecCommandLoop.
*/
type ConnectionState interface {
	Name() string
	// Init gets called to set up any information in UserConnection
	// required by respond
	Init(*UserConnection)
	// Respond is called on user input. If false is returned, the 
	// connection leaves out of band mode.
	Respond(*UserConnection) bool
}

type UndefinedState struct {
	ConnectionState
}

func (s *UndefinedState) Name() string { return "Undefined placeholder" }
func (s *UndefinedState) Init(c *UserConnection) {
	c.Write("Connection state undefined, contact admin.\n\r")
}
func (s *UndefinedState) Respond(c *UserConnection) bool {
	c.Close()
	return true
}

func init() {
	colorMap = make(map[string]string)
	colorMap["&dim;"] = "\x1b[2m"
	colorMap["&black;"] = "\x1b[30m"
	colorMap["&red;"] = "\x1b[31m"
	colorMap["&green;"] = "\x1b[32m"
	colorMap["&yellow;"] = "\x1b[33m"
	colorMap["&blue;"] = "\x1b[34m"
	colorMap["&magenta;"] = "\x1b[35m"
	colorMap["&cyan;"] = "\x1b[36m"
	colorMap["&white;"] = "\x1b[37m"
	colorMap["&;"] = "\x1b[0m"
}

/* 
 Opens up a new UserConnection connected by socket in state connectState.

 Initiates the read loop which will populate the user I/O (FromUser and 
 ToUser channels)
*/
func NewUserConnection(socket net.Conn, connectState ConnectionState) *UserConnection {
	c := new(UserConnection)
	c.socket = socket
	c.State = connectState
	c.FromUser = make(chan string, 10)
	c.ToUser = make(chan string, 10)
	c.done = make(chan bool, 1)
	c.outOfBand = true
	c.Data = make(map[string]interface{})
	
	go c.readLoop()
	return c
}


func (c *UserConnection) Close() { 
	c.done <- true
}

func (c *UserConnection) Write(text string) { 	
	str_acc := text
	for easyCode, termCode := range colorMap {
		str_acc = strings.Replace(str_acc, easyCode, termCode, -1)
	}
	c.socket.Write([]byte(str_acc))
}

func (c *UserConnection) readLoop() {
	rawBuf := make([]byte, 1024)
	defer c.socket.Close()

	c.State.Init(c)
	for {
		select {
		case <-c.done:
			c.OnDisconnect()
			return
		default:
			n, err := c.socket.Read(rawBuf)
			if err == nil {
				strCommand := string(rawBuf[:n])
				c.FromUser <- strings.TrimRight(strCommand,"\n\r")
			} else {
				c.done <- true
			}
			if(c.outOfBand) {
				c.outOfBand = c.State.Respond(c)
			}
		}
	}
}