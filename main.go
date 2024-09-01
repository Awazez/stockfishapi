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

	// Récupérer le paramètre depth (profondeur), avec une valeur par défaut de 15
	depth := r.URL.Query().Get("depth")
	if depth == "" {
		depth = "15"
	}

	// Récupérer le paramètre threads, avec une valeur par défaut de 1
	threads := r.URL.Query().Get("threads")
	if threads == "" {
		threads = "1"
	}

	// Créer la commande Stockfish
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

	// Démarrer le processus Stockfish
	if err := cmd.Start(); err != nil {
		http.Error(w, "Failed to start Stockfish", http.StatusInternalServerError)
		return
	}

	// Envoyer les commandes à Stockfish
	go func() {
		defer stdin.Close()
		fmt.Fprintln(stdin, "uci")
		fmt.Fprintln(stdin, "setoption name Threads value", threads) // Configurer le nombre de threads
		fmt.Fprintln(stdin, "position fen", fen)
		fmt.Fprintln(stdin, "go depth", depth) // Configurer la profondeur de recherche
	}()

	// Capturer la sortie
	var buf strings.Builder
	go func() {
		defer stdout.Close()
		io.Copy(&buf, stdout)
	}()

	// Attendre la fin du processus
	if err := cmd.Wait(); err != nil {
		http.Error(w, "Stockfish process failed", http.StatusInternalServerError)
		return
	}

	// Extraire le meilleur coup de la sortie
	output := buf.String()
	bestMove := extractBestMove(output)
	if bestMove == "" {
		http.Error(w, "Failed to find best move", http.StatusInternalServerError)
		return
	}

	// Retourner le meilleur coup en réponse
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
