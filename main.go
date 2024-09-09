package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
)

func getBestMove(w http.ResponseWriter, r *http.Request) {
	fen := r.URL.Query().Get("fen")
	if fen == "" {
		http.Error(w, "Missing FEN parameter", http.StatusBadRequest)
		return
	}

	depth := r.URL.Query().Get("depth")
	if depth == "" {
		depth = "15" // Default depth
	} else {
		if _, err := strconv.Atoi(depth); err != nil {
			http.Error(w, "Invalid depth parameter", http.StatusBadRequest)
			return
		}
	}

	threads := r.URL.Query().Get("threads")
	if threads == "" {
		threads = "1" // Default threads
	} else {
		if _, err := strconv.Atoi(threads); err != nil {
			http.Error(w, "Invalid threads parameter", http.StatusBadRequest)
			return
		}
	}

	cmd := exec.Command("stockfish")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		http.Error(w, "Failed to create stdin pipe", http.StatusInternalServerError)
		return
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		http.Error(w, "Failed to create stdout pipe", http.StatusInternalServerError)
		return
	}

	if err := cmd.Start(); err != nil {
		http.Error(w, "Failed to start Stockfish", http.StatusInternalServerError)
		return
	}

	go func() {
		defer stdin.Close()
		fmt.Fprintln(stdin, "uci")
		fmt.Fprintf(stdin, "setoption name Threads value %s\n", threads)
		fmt.Fprintf(stdin, "position fen %s\n", fen)
		fmt.Fprintf(stdin, "go depth %s\n", depth)
	}()

	scanner := bufio.NewScanner(stdout)
	var response strings.Builder

	for scanner.Scan() {
		line := scanner.Text()
		// Capture and append all lines related to engine's thinking process
		response.WriteString(line + "\n")

		// Stop reading after the best move is found
		if strings.HasPrefix(line, "bestmove") {
			break
		}
	}

	if err := cmd.Wait(); err != nil {
		http.Error(w, "Stockfish process failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(response.String()))
}

func main() {
	http.HandleFunc("/bestmove", getBestMove)
	fmt.Println("Server started at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
