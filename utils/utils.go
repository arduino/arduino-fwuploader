package utils

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type firmware struct {
	Path     string
	Name     string
	IsLoader bool
}

type combo struct {
	match  string
	prefer string
	avoid  string
	loader string
}

func isPreferred(existing bool, path string, board combo) bool {
	if path == "" {
		return false
	}
	if board.avoid != "" && strings.Contains(path, board.avoid) {
		return false
	}
	if existing && !strings.Contains(path, board.prefer) {
		return false
	}
	return true
}

func GetCompatibleWith(name string) map[string][]firmware {

	files := make(map[string][]firmware)

	knownBoards := make(map[string]combo)
	knownBoards["mkr1000"] = combo{match: "(WINC1500)*(3a0)", loader: "WINC1500/Firmware*"}
	knownBoards["mkrwifi1010"] = combo{match: "(NINA)", loader: "NINA/Firmware*(mkrwifi)*", avoid: "uno"}
	knownBoards["mkrvidor4000"] = combo{match: "(NINA)", loader: "NINA/Firmware*(mkrvidor)*", avoid: "uno"}
	knownBoards["uno2018"] = combo{match: "(NINA)", loader: "NINA/Firmware*(unowifi)*", prefer: "uno", avoid: "mkr"}
	knownBoards["mkrnb1500"] = combo{match: "SARA", loader: "SARA/SerialSARAPassthrough*"}

	listAll := false

	if knownBoards[strings.ToLower(name)].match == "" {
		listAll = true
	}

	exePath, _ := os.Executable()
	root := filepath.Dir(exePath)
	root = filepath.Join(root, "..", "firmwares")

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		unixPath := filepath.ToSlash(path)
		parts := strings.Split(unixPath, "/")
		fancyName := parts[len(parts)-3] + " " + parts[len(parts)-2]
		loader := regexp.MustCompile(knownBoards[name].loader)
		fw := regexp.MustCompile(knownBoards[name].match)
		f := firmware{
			Path:     path,
			Name:     fancyName,
			IsLoader: loader.MatchString(path) && !listAll,
		}
		if info.IsDir() {
			return nil
		}
		folder := filepath.Dir(path)
		lowerPath, _ := filepath.Rel(root, path)
		lowerPath = strings.ToLower(lowerPath)
		_, alreadyPopulated := files[folder]
		if listAll || ((fw.MatchString(path) || f.IsLoader) && isPreferred(alreadyPopulated, lowerPath, knownBoards[name])) {
			files[folder] = append(files[folder], f)
		}
		return nil
	})
	if err != nil {
		return files
	}
	return files
}
