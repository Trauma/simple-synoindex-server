package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/go-ini/ini"
)

var (
	inifile        string
	cfg            *ini.File
	volumeMappings map[string]string
	lastMTime      time.Time
)

func init() {

	// get current execute file path
	execdir := GetCurrentExecDir()

	inifile = fmt.Sprintf("%s/simple-synoindex-server.ini", execdir)
	cfg, _ = ini.LooseLoad(inifile)

	reloadMappings()

}

func reloadMappings() {

	stat, err := os.Stat(inifile)

	if err != nil {
		log.Printf("reloadMappings Error: %s \n", err)
		return
	}

	iniMTime := stat.ModTime()

	if iniMTime.After(lastMTime) {
		cfg.Reload()
		volumeMappings = cfg.Section("mappings").KeysHash()
		lastMTime = iniMTime
	}

}

func remappingPath(srcPath string) string {

	newPath := srcPath

	for vPath, mPath := range volumeMappings {
		newPath = strings.Replace(newPath, vPath, mPath, 1)
		if newPath != srcPath {
			return newPath
		}
	}

	return newPath
}

func SynoIndex(w http.ResponseWriter, req *http.Request) {

	io.WriteString(w, "ok\n")
	req.ParseForm()
	args := req.Form["args"]

	// skip execute synoindex if only one argument
	if len(args) == 1 {
		log.Printf("Simple-SynoIndex NOT Support [%s] argument, but response OK to clients!\n", args[0])
		return
	}

	// reload mapping settings if necessarily
	reloadMappings()

	args[1] = remappingPath(args[1])
	args[2] = remappingPath(args[2])

	// log to stdout
	log.Printf("SynoIndex: %s %s %s \n", args[0], args[1], args[2])

	// execute /usr/syno/bin/synoindex
	cmd := exec.Command("/usr/syno/bin/synoindex", args...)

	err := cmd.Run()

	if err != nil {
		log.Printf("SynoIndex Error: %s \n")
	}

}

func main() {

	srvIp := cfg.Section("main").Key("SERVER_IP").MustString("172.17.0.1")
	srvPort := cfg.Section("main").Key("SERVER_PORT").MustString("32699")
	srvListen := fmt.Sprintf("%s:%s", srvIp, srvPort)

	http.HandleFunc("/synoindex", SynoIndex)
	log.Fatal(http.ListenAndServe(srvListen, nil))

}
