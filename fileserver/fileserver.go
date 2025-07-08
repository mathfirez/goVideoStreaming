package fileserver

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

// Starts file server.
func FileServerStart() {
	port := os.Getenv("API_PORT")
	port = ":" + port
	fileServer := http.FileServer(http.Dir("./files"))

	http.Handle("/files/", http.StripPrefix("/files/", fileServer))

	fmt.Println("-- Starting server: http://localhost", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

// Starts ngrok.
func NgrokStart() {
	port := os.Getenv("API_PORT")
	fmt.Println("-- Running NGROK...")
	ngrokExePath := os.Getenv("NGROK_EXE_PATH")
	cmd := exec.Command(ngrokExePath+"ngrok.exe", "http", port)

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("NGROK Error:\n%s\n%v", string(output), err)
	}

}

type NgrokTunnel struct {
	PublicURL string `json:"public_url"`
}

type NgrokAPIResponse struct {
	Tunnels []NgrokTunnel `json:"tunnels"`
}

// Grabs the main url.
func NgrokGetURL() string {
	ngrokTunnelURL := os.Getenv("NGROK_TUNNEL_URL")
	resp, err := http.Get(ngrokTunnelURL)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var result NgrokAPIResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		panic(err)
	}

	for _, tunnel := range result.Tunnels {
		if strings.HasPrefix(tunnel.PublicURL, "https://") {
			return tunnel.PublicURL
		}
	}
	panic("Ngrok tunnel not found")
}
