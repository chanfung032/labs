package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"sync"
	"text/template"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var (
	blockIpScriptTpl = template.Must(template.New("block").Parse(`
ipset create {{.Domain}} hash:ip hashsize 819200 maxelem 100000 timeout 300 -exist
for ip in {{range $ip := .Ips}} {{$ip}} {{end}}; do
	ipset add {{.Domain}} $ip timeout {{.Timeout}} -exist
done
[ $(iptables -L |grep '"{{.Domain}}"' | wc -l) -eq 0 ] &&
	iptables -A INPUT -p tcp -m tcp -m multiport --dports 80,443 -m set --match-set {{.Domain}} src -m string --string {{.Domain}} --algo kmp --to 1480 -j NFQUEUE --queue-num 0 -m comment --comment "b10ck"
{{.TcpForceResetExecutable}} 0
`))

	cleanIpScriptTpl = template.Must(template.New("clean").Parse(`
iptables -L --line-numbers | grep '"{{.Domain}}"' | while read line; do
	iptables -D INPUT $(echo $line | awk '{print $1}')
done
ipset destroy {{.Domain}}
`))

	tcpForceResetExecutable = fmt.Sprintf("%s/tcp_force_reset", path.Dir(os.Args[0]))

	mutex = &sync.Mutex{}
)

func block(message []byte) {
	var blockRequest struct {
		Domain  string
		Timeout int
		Ips     []string
	}
	err := json.Unmarshal(message, &blockRequest)
	if err != nil {
		log.Error(err)
		return
	}
	log.WithField("request", blockRequest).Info("got request")

	context := map[string]interface{}{
		"Domain":                  blockRequest.Domain,
		"Timeout":                 blockRequest.Timeout,
		"Ips":                     blockRequest.Ips,
		"TcpForceResetExecutable": tcpForceResetExecutable,
	}
	var b bytes.Buffer
	if err := blockIpScriptTpl.Execute(&b, context); err != nil {
		log.Errorf("render script failed: %s", err)
		return
	}
	script := b.String()

	log.WithField("script", script).Info("block")
	mutex.Lock()
	defer mutex.Unlock()
	if _, err := sh(script); err != nil {
		log.Error(err)
	}
}

func clean() {
	mutex.Lock()
	defer mutex.Unlock()

	output, err := sh("ipset -l -output xml")
	if err != nil {
		log.Error(err)
		return
	}

	decoder := xml.NewDecoder(bytes.NewBuffer(output))
	for {
		var ipset struct {
			Domain  string   `xml:"name,attr"`
			Members []string `xml:"members>member"`
		}
		err := decoder.Decode(&ipset)
		if err != nil {
			if err != io.EOF {
				log.Error(err)
			}
			break
		}
		if len(ipset.Members) != 0 {
			continue
		}

		var b bytes.Buffer
		if err := cleanIpScriptTpl.Execute(&b, &ipset); err != nil {
			log.Error(err)
			continue
		}
		log.WithField("script", b.String()).Info("clean")

		if _, err := sh(b.String()); err != nil {
			log.Error(err)
		}
	}
}

func sh(cmd string) ([]byte, error) {
	return exec.Command("sh", "-c", cmd).CombinedOutput()
}

func agent(c *cli.Context) {
	go func() {
		for {
			clean()
			time.Sleep(1 * time.Minute)
		}
	}()

	if c.String("master") != "" {
		u := url.URL{Scheme: "ws", Host: c.String("master"), Path: "/ws"}
		log.Infof("master is %s", u.String())

		for {
			conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
			if err != nil {
				log.Errorf("dial: %s", err)
				time.Sleep(5 * time.Second)
				continue
			}
			for {
				_, message, err := conn.ReadMessage()
				if err != nil {
					log.Error(err)
					break
				}
				go block(message)
			}
			conn.Close()
		}
	} else {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				log.Error(err)
				return
			}
			block(body)
		})

		log.Fatal(http.ListenAndServe(c.String("addr"), nil))
	}
}
