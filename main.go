package main

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"
)

func bestMove(w http.ResponseWriter, r *http.Request) {
	position := r.URL.Query().Get("position")
	if position == "" {
		http.Error(w, "Missing 'position' query parameter", http.StatusBadRequest)
		return
	}

	// Lancer Stockfish en tant que processus
	cmd := exec.Command("stockfish")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		http.Error(w, "Failed to get stdin for Stockfish", http.StatusInternalServerError)
		return
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		http.Error(w, "Failed to get stdout for Stockfish", http.StatusInternalServerError)
		return
	}

	// Démarrer Stockfish
	if err := cmd.Start(); err != nil {
		http.Error(w, "Failed to start Stockfish", http.StatusInternalServerError)
		return
	}

	// Envoyer la position et demander le meilleur coup
	go func() {
		fmt.Fprintln(stdin, "uci")
		fmt.Fprintln(stdin, "position fen", position)
		fmt.Fprintln(stdin, "go movetime 1000") // 1 seconde pour trouver le meilleur coup
		stdin.Close()
	}()

	// Lire la sortie de Stockfish
	buf := new(strings.Builder)
	go func() {
		buf.ReadFrom(stdout)
	}()

	cmd.Wait()

	// Extraire le meilleur coup de la sortie de Stockfish
	output := buf.String()
	move := extractBestMove(output)

	if move == "" {
		http.Error(w, "Failed to find best move", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"best_move": "%s"}`, move)
}

func extractBestMove(output string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "bestmove") {
			parts := strings.Fields(line)
			if len(parts) > 1 {
				return parts[1]
			}
		}
	}
	return ""
}

func main() {
	http.HandleFunc("/bestmove", bestMove)
	log.Println("API is running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
