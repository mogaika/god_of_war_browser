package utils

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
)

func GameStringHashNodes(str string, initial uint32) uint32 {
	hash := initial
	for _, c := range str {
		hash = hash*127 + uint32(byte(c))
	}
	return hash
}

var hashesMap sync.Map

func loadHashes(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	r := bufio.NewReader(f)
	for {
		line, err := r.ReadString('\n')
		if err == io.EOF {
			return nil
		}

		if strings.HasPrefix(line, "#") {
			continue
		}

		line = strings.TrimSuffix(line, "\n")
		line = strings.TrimSuffix(line, "\r")
		var hash, init uint32
		n, err := fmt.Sscanf(line, "%x:%x:", &hash, &init)
		if n != 2 {
			continue
		}
		if err != nil {
			return err
		}
		parts := strings.SplitN(line, ":", 3)
		input := parts[2]

		if hash != GameStringHashNodes(input, init) {
			log.Printf("!!! Invalid hash in hashes txt (%q)", input)
		}

		if init == 0 {
			// TODO: store init != 0 too
			hashesMap.Store(hash, input)
		}
	}
}

func loadStringHashes(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	r := bufio.NewReader(f)
	for {
		line, err := r.ReadString('\n')
		if err == io.EOF {
			return nil
		}

		if strings.HasPrefix(line, "#") {
			continue
		}

		line = strings.TrimSuffix(line, "\n")
		line = strings.TrimSuffix(line, "\r")

		hashesMap.Store(GameStringHashNodes(line, 0), line)
	}
}

func init() {
	go func() {
		if err := loadHashes("hashes.dump.txt"); err != nil {
			log.Printf("Failed to load hash file: %v", err)
		}

		if err := loadHashes("strings.dump.txt"); err != nil {
			log.Printf("Failed to load string hashes file: %v", err)
		}
	}()
}

func GameStringUnhashGenerate(hash uint32) string {
	s := ""
	for hash != 0 {
		s += string(rune(hash % 127))
		hash /= 127
	}
	return ReverseString(s)
}

func GameStringUnhashNodes(hash uint32) string {
	v, ok := hashesMap.Load(hash)
	if ok {
		return v.(string)
	} else {
		return "%gene% " + GameStringUnhashGenerate(hash)
	}
}
