package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"strings"
)

func getBestMove(w http.ResponseWriter, r *http.Request) {
	fen := r.URL.Query().Get("fen")
	if fen == "" {
		http.Error(w, "Missing FEN parameter", http.StatusBadRequest)
		return
	}

	// Create the Stockfish command
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

	// Start the Stockfish process
	if err := cmd.Start(); err != nil {
		http.Error(w, "Failed to start Stockfish", http.StatusInternalServerError)
		return
	}

	// Send commands to Stockfish
	go func() {
		defer stdin.Close()
		fmt.Fprintln(stdin, "uci")
		fmt.Fprintln(stdin, "position fen", fen)
		fmt.Fprintln(stdin, "go depth 15")
	}()

	// Capture the output
	var buf strings.Builder
	go func() {
		defer stdout.Close()
		io.Copy(&buf, stdout)
	}()

	// Wait for the process to finish
	if err := cmd.Wait(); err != nil {
		http.Error(w, "Stockfish process failed", http.StatusInternalServerError)
		return
	}

	// Find the best move in the output
	output := buf.String()
	bestMove := extractBestMove(output)
	if bestMove == "" {
		http.Error(w, "Failed to find best move", http.StatusInternalServerError)
		return
	}

	// Write the best move to the response
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(bestMove))
}

func extractBestMove(output string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "bestmove") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				return parts[1]
			}
		}
	}
	return ""
}

func main() {
	http.HandleFunc("/bestmove", getBestMove)
	fmt.Println("Server started at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
