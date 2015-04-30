package dbmaker

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"streamdb/util"

)

var (
	//ErrTimeout is thrown when a port does not open on time
	ErrTimeout = errors.New("Timeout on operation reached - it looks like something crashed.")
	//ErrProcessNotFound is thrown when the process is not found
	ErrProcessNotFound = errors.New("The process was not found")

	//PortTimeoutLoops go in time of 100 milliseconds
	PortTimeoutLoops = 100
)

//WaitPort waits for a port to open
func WaitPort(host string, port int, err error) error {
	if err != nil {
		return err
	}

	hostPort := fmt.Sprintf("%s:%d", host, port)

	log.Printf("Waiting for %v to open...\n", hostPort)

	_, err = net.Dial("tcp", hostPort)
	i := 0
	for ; err != nil && i < PortTimeoutLoops; i++ {
		_, err = net.Dial("tcp", hostPort)
		time.Sleep(100 * time.Millisecond)
	}
	if i >= PortTimeoutLoops {
		return ErrTimeout
	}

	log.Printf("...%v is now open.\n", hostPort)
	return nil
}

func cmd2Str(command string, args ...string) string {
	return fmt.Sprintf("> %v %v", command, strings.Join(args, " "))
}

//RunCommand runs the given command in foreground
func RunCommand(err error, command string, args ...string) error {
	if err != nil {
		return err
	}
	log.Printf(cmd2Str(command, args...))

	cmd := exec.Command(command, args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

//RunDaemon runs the given command as a daemon (in the background)
func RunDaemon(err error, command string, args ...string) error {
	if err != nil {
		return err
	}
	log.Printf(cmd2Str(command, args...))

	cmd := exec.Command(command, args...)

	//No need for redirecting stuff, since log/pid files are configured in .conf files
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	//I am not convinced at the moment that restarting postgres/other stuff will be a good idea
	//especially since that is what happens when we want to kill them from another process.
	//So, for the moment, just start the process
	return cmd.Start()
}

//GetProcess gets the gven process using its process name
func GetProcess(streamdbDirectory, procname string, err error) (*os.Process, error) {
	if err != nil {
		return nil, err
	}

	pidfile := filepath.Join(streamdbDirectory, procname+".pid")
	if !util.PathExists(pidfile) {
		return nil, ErrFileNotFound
	}

	pidbytes, err := ioutil.ReadFile(pidfile)
	if err != nil {
		return nil, err
	}

	pids := strings.Fields(string(pidbytes))

	if len(pids) < 1 {
		return nil, ErrProcessNotFound
	}

	pid, err := strconv.Atoi(pids[0])
	if err != nil {
		return nil, err
	}

	return os.FindProcess(pid)
}

//StopProcess sends SIGINT to the process - NOTE: as of writing, godocs says sigint not availalbe on windows
func StopProcess(streamdbDirectory, procname string, err error) error {
	p, err := GetProcess(streamdbDirectory, procname, err)
	if err != nil {
		return err
	}
	return p.Signal(os.Interrupt)
}

//KillProcess sends immediately kills the process
func KillProcess(streamdbDirectory, procname string, err error) error {
	p, err := GetProcess(streamdbDirectory, procname, err)
	if err != nil {
		return err
	}
	return p.Kill()
}
