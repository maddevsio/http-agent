package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
	"github.com/maddevsio/http-agent/agent"
)

type Counters struct {
	TargetHostname    string `json:"target_hostname"`
	TimeStartTransfer string `json:"response_time_start_transfer"`
	ResponseTimeTotal string `json:"response_time_total"`
}

const (
	defaultPort         = "8090"
	defaultDashboardURL = "http://localhost:8080/dashboard/v1/register"
	defaultTarget       = "google.com:80"
	defaultIPAddress    = "127.0.0.1"
)

func main() {
	var (
		addr   = envString("PORT", defaultPort)
		durl   = envString("DASHBOARD_URL", defaultDashboardURL)
		tHost  = envString("TARGET_HOST", defaultTarget)
		ipAddr = envString("IP_ADDRESS", defaultIPAddress)

		httpAddr       = flag.String("httpAddr", "0.0.0.0:"+addr, "HTTP listen address")
		dashboardURL   = flag.String("dashboardURL", durl, "Dashboard service URL")
		targetHostname = flag.String("targetHost", tHost, "Target hostname and port")
		listenIPAddr   = flag.String("ipAddr", ipAddr, "HTTP listen ip address")
	)

	flag.Parse()

	err := agent.Register(*dashboardURL, "http://"+*listenIPAddr+":"+addr)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Http listen address: %s", *httpAddr)

	e := echo.New()
	e.File("/", "tmpl/index.html")
	e.GET("/check", func(c echo.Context) error {
		conn, err := net.Dial("tcp", *targetHostname)
		if err != nil {
			return fmt.Errorf("dial error:", err)
		}
		defer conn.Close()
		fmt.Fprintf(conn, "GET / HTTP/1.0\r\n\r\n")

		startTime := time.Now()
		oneByte := make([]byte, 1)
		for {
			_, err = conn.Read(oneByte)
			log.Println("read from buffer")
			if err != nil {
				if err != io.EOF {
					return fmt.Errorf("read error:", err)
				}
				break
			}
		}
		timeStart := time.Since(startTime).String()

		_, err = ioutil.ReadAll(conn)
		if err != nil {
			log.Println(err)
		}
		responseTotal := time.Since(startTime).String()
		return c.JSON(http.StatusOK, Counters{
			TargetHostname:    *targetHostname,
			TimeStartTransfer: timeStart,
			ResponseTimeTotal: responseTotal,
		})
	})
	e.Run(standard.New(*httpAddr))
}

func envString(env, fallback string) string {
	e := os.Getenv(env)
	if e == "" {
		return fallback
	}
	return e
}
