package main

import (
	"bufio"
	"fmt"
	"hash/fnv"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pterm/pterm"
	"github.com/pterm/pterm/putils"

	"github.com/bominrahmani/kino/providers"
)

type searchResultMsg struct {
	movies []*providers.Movie
	err    error
}

func main() {
	introductionMessage()

	// Prompt user for movie name
	var movieName string
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter a movie name: ")
	movieName, _ = reader.ReadString('\n')
	movieName = strings.TrimSpace(movieName)

	// Feed movie into engine
	flixhq := providers.NewFlixHQProvider()
	catalogue, err := flixhq.Scrape(movieName)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Download and cache all the images within a temporary directory
	downloadImages(catalogue)

	// Use FZF to select a movie
	selectedMovie, err := FZFSearch(catalogue)
	if err != nil {
		fmt.Println("Error selecting movie:", err)
		return
	}

	fmt.Println(selectedMovie)

}

func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

func upscaleImage(url string) string {
	re := regexp.MustCompile(`/\d+x\d+/`)
	return re.ReplaceAllString(url, "/1000x1000/")
}

func downloadImages(catalogue []*providers.Movie) {
	for _, movie := range catalogue {
		// upscale image link
		fmt.Println("UPSCALED: ", upscaleImage(movie.ImageUrl))
		upscaledImageUrl := upscaleImage(movie.ImageUrl)
		resp, err := http.Get(upscaledImageUrl)
		if err != nil {
			fmt.Println("Error downloading image previews: ", err)
			return
		}

		defer resp.Body.Close()

		imageHash := hash(movie.ImageUrl)
		imageFile := filepath.Join("/tmp/kinoImages/", fmt.Sprint(imageHash))
		out, err := os.Create(imageFile)

		if err != nil {
			fmt.Println("Error downloading image previews: ", err)
			return
		}
		defer out.Close()

		_, err = io.Copy(out, resp.Body)
		if err != nil {
			fmt.Println("Error downloading image previews: ", err)
			return
		}
	}
}

func FZFSearch(catalogue []*providers.Movie) (*providers.Movie, error) {
	var input strings.Builder
	for _, movie := range catalogue {
		hashedFileName := fmt.Sprint(hash(movie.ImageUrl))
		fullImagePath := filepath.Join("/tmp/kinoImages/", hashedFileName)
		input.WriteString(fmt.Sprintf("%s (%s)\t%s\n", movie.Title, movie.Year, fullImagePath))
	}

	// Calculate image dimensions and position
	imageWidth := 50
	imageHeight := 50 
	xOffset := 10 
	yOffset := 0

	previewCmd := fmt.Sprintf("kitty +kitten icat --clear --place %dx%d@%dx%d --scale-up  --stdin=no --transfer-mode file {2}",
		imageWidth, imageHeight, xOffset, yOffset)


	fzfArgs := []string{
    "--cycle",
    "--reverse",
		"--with-nth", "1",
		"-d", "\t",
		"--preview", previewCmd,
    "--preview-window", "noborder",
		"--preview-window", "right:40%",
	}
  cmd := exec.Command("fzf", fzfArgs...)
	cmd.Stdin = strings.NewReader(input.String())
	cmd.Stderr = os.Stderr

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	selectedTitle := strings.TrimSpace(string(output))
	for _, movie := range catalogue {
		if fmt.Sprintf("%s (%s)", movie.Title, movie.Year) == selectedTitle {
			return movie, nil
		}
	}

	return nil, fmt.Errorf("selected movie not found in catalogue")
}

func introductionMessage() {
	fmt.Print("\033[H\033[2J")
	s, _ := pterm.DefaultBigText.WithLetters(putils.LettersFromStringWithStyle("KINO", pterm.FgCyan.ToStyle())).Srender()
	pterm.DefaultCenter.Println(s)
	pterm.DefaultCenter.WithCenterEachLineSeparately().Println("Totus Mundus\nAgit Histrionem")

}
