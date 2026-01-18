package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type SSHExecRequest struct {
	Command string `json:"command"`
}

type SSHExecResponse struct {
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode int    `json:"exit_code"`
	Error    string `json:"error,omitempty"`
}

const BaseURL = "http://111.92.242.223:48891/api/ssh/exec"

func execute(cmd string) {
	fmt.Printf(">>> Executing: %s\n", cmd)
	encodedCmd := base64.StdEncoding.EncodeToString([]byte(cmd))

	reqBody := SSHExecRequest{Command: encodedCmd}
	jsonData, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", BaseURL, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("!!! Request Failed: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		fmt.Printf("!!! HTTP Error: %s\n%s\n", resp.Status, string(body))
		return
	}

	var execResp SSHExecResponse
	if err := json.Unmarshal(body, &execResp); err != nil {
		fmt.Printf("!!! JSON Parse Error: %v\n%s\n", err, string(body))
		return
	}

	decodedStdout, _ := base64.StdEncoding.DecodeString(execResp.Stdout)
	decodedStderr, _ := base64.StdEncoding.DecodeString(execResp.Stderr)

	if execResp.ExitCode != 0 {
		fmt.Printf("!!! Remote Execution Failed (Code %d)\n", execResp.ExitCode)
		if len(decodedStderr) > 0 {
			fmt.Printf("Stderr: %s\n", decodedStderr)
		}
		if execResp.Error != "" {
			decodedErr, _ := base64.StdEncoding.DecodeString(execResp.Error)
			fmt.Printf("API Error: %s\n", decodedErr)
		}
	} else {
		fmt.Printf("SUCCESS.\n")
		fmt.Printf("Output: %s\n", decodedStdout) // Print output
	}
}

func main() {
	fmt.Println("Deploying vsftpd using AI_SSH_FTP proxy...")

	// 1. Update APT
	execute("apt-get update")

	// 2. Install vsftpd
	execute("apt-get install -y vsftpd")

	// 3. Enable Write in config (sed)
	// Default vsftpd.conf usually has #write_enable=YES. We uncomment it.
	execute("sed -i 's/#write_enable=YES/write_enable=YES/' /etc/vsftpd.conf")

	// 4. Restart Service
	execute("systemctl restart vsftpd")
	execute("systemctl enable vsftpd")

	// 5. Verify Port 21
	execute("ss -tuln | grep :21")

	fmt.Println("Deployment sequence completed.")
}
