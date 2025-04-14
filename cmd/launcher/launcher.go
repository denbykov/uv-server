package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

func main() {
	err := updateIfNeeded()
	if err != nil {
		fmt.Printf("error: %v", err)
		return
	}

	err = start()
	if err != nil {
		fmt.Printf("error: %v", err)
		return
	}
}

func start() error {
	fmt.Println("starting...")

	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	err = startServer(wd)
	if err != nil {
		return err
	}

	err = startClient(wd)
	if err != nil {
		return err
	}

	fmt.Println("starting completed")

	return nil
}

func startServer(wd string) error {
	path := path.Join(wd, "server", "uv_server")
	cmd := exec.Command(path)
	cmd.Dir = "server"
	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to run server: %v", err)
	}

	return nil
}

func startClient(wd string) error {
	path := path.Join(wd, "client", "uv-client")
	cmd := exec.Command(path)
	cmd.Dir = "client"
	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to run client: %v", err)
	}

	return nil
}

func updateIfNeeded() error {
	fmt.Println("checking for updates")
	serverTagsUrl := "https://api.github.com/repos/denbykov/uv-server/tags"
	serverVersion, err := getLatestVersion(serverTagsUrl)
	if err != nil {
		return fmt.Errorf("failed to get latest server version: %v", err)
	}
	fmt.Printf("latest server version: %v\n", serverVersion)

	clientTagsUrl := "https://api.github.com/repos/denbykov/uv-client/tags"
	clientVersion, err := getLatestVersion(clientTagsUrl)
	if err != nil {
		return fmt.Errorf("failed to get latest client version: %v", err)
	}
	fmt.Printf("latest client version: %v\n", clientVersion)

	versionFileName := "version"

	currentVersion, err := readVersion(versionFileName)

	if err == nil {
		fmt.Printf("current version: %v\n", currentVersion)

		if compareSemVer(serverVersion, clientVersion) == 0 &&
			compareSemVer(currentVersion, serverVersion) == -1 {
			fmt.Printf("updating to %v\n", serverVersion)

			err := installLatestVersion(serverVersion)
			if err != nil {
				return err
			}

			err = writeVersion(versionFileName, serverVersion)
			if err != nil {
				return err
			}

			fmt.Println("update completed")
		}
	} else {
		fmt.Println("current version is not set")

		if compareSemVer(serverVersion, clientVersion) == 0 {
			fmt.Printf("updating to %v\n", serverVersion)

			err := installLatestVersion(serverVersion)
			if err != nil {
				return err
			}

			err = writeVersion(versionFileName, serverVersion)
			if err != nil {
				return err
			}

			fmt.Println("update completed")
		} else {
			return fmt.Errorf("unable to install latest version, client and server version mismatch")
		}
	}

	return nil
}

func installLatestVersion(version string) error {
	err := installServer(version)
	if err != nil {
		return err
	}

	return installClient(version)
}

func writeVersion(filename, version string) error {
	return os.WriteFile(filename, []byte(version), 0644)
}

func readVersion(filename string) (string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func compareSemVer(v1, v2 string) int {
	split1 := strings.Split(v1, ".")
	split2 := strings.Split(v2, ".")

	for i := 0; i < 3; i++ {
		n1, _ := strconv.Atoi(split1[i])
		n2, _ := strconv.Atoi(split2[i])

		if n1 > n2 {
			return 1
		} else if n1 < n2 {
			return -1
		}
	}
	return 0
}

type Commit struct {
	SHA string `json:"sha"`
	URL string `json:"url"`
}
type Version struct {
	Name       string `json:"name"`
	ZipballURL string `json:"zipball_url"`
	TarballURL string `json:"tarball_url"`
	Commit     Commit `json:"commit"`
	NodeID     string `json:"node_id"`
}

func getLatestVersion(url string) (tag string, err error) {
	tags, err := getTags(url)
	if err != nil {
		return tag, err
	}

	tag = tags[0].Name
	return tag, err
}

func getTags(url string) (tags []Version, err error) {
	resp, err := http.Get(url)
	if err != nil {
		return tags, err
	}
	defer resp.Body.Close()

	var buf bytes.Buffer

	_, err = io.Copy(&buf, resp.Body)
	if err != nil {
		return tags, err
	}

	err = json.Unmarshal(buf.Bytes(), &tags)

	return tags, err
}

func installServer(tag string) error {
	url := fmt.Sprintf("https://github.com/denbykov/uv-server/releases/download/%v/server.zip", tag)
	err := downloadFile(url, "server.zip")
	if err != nil {
		return err
	}

	err = unzipSource("server.zip", ".")
	if err != nil {
		return err
	}

	err = os.Remove("server.zip")
	return err
}

func installClient(tag string) error {
	url := fmt.Sprintf("https://github.com/denbykov/uv-client/releases/download/%v/client.zip", tag)
	err := downloadFile(url, "client.zip")
	if err != nil {
		return err
	}

	err = unzipSource("client.zip", ".")
	if err != nil {
		return err
	}

	err = os.Remove("client.zip")
	return err
}

func downloadFile(url, destination string) error {
	fmt.Printf("downloading file from url: %v\n", url)
	fmt.Printf("destination is: %v\n", destination)

	if info, err := os.Stat(destination); err == nil {
		if info.IsDir() {
			return fmt.Errorf("requested path is a directory")
		}

		if err = os.Remove(destination); err != nil {
			return err
		}
	}

	out, err := os.Create(destination)

	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	size, _ := strconv.Atoi(resp.Header.Get("Content-Length"))

	reader := &ProgressReader{
		Reader: resp.Body,
		Total:  int64(size),
	}

	_, err = io.Copy(out, reader)

	// extra spaces to remove potential garbage from progess display
	fmt.Printf("donwloading done             ")
	return err
}

type ProgressReader struct {
	Reader io.Reader
	Total  int64
	Count  int64
}

func (p *ProgressReader) Read(b []byte) (int, error) {
	n, err := p.Reader.Read(b)
	p.Count += int64(n)
	fmt.Printf("Downloading... %d%%", (p.Count*100)/p.Total)
	fmt.Printf("\r")
	return n, err
}

func unzipSource(source, destination string) error {
	reader, err := zip.OpenReader(source)
	if err != nil {
		return err
	}
	defer reader.Close()

	destination, err = filepath.Abs(destination)
	if err != nil {
		return err
	}

	for _, f := range reader.File {
		err := unzipFile(f, destination)
		if err != nil {
			return err
		}
	}

	return nil
}

func unzipFile(f *zip.File, destination string) error {
	filePath := filepath.Join(destination, f.Name)
	if !strings.HasPrefix(filePath, filepath.Clean(destination)+string(os.PathSeparator)) {
		return fmt.Errorf("invalid file path: %s", filePath)
	}

	if f.FileInfo().IsDir() {
		if err := os.MkdirAll(filePath, os.ModePerm); err != nil {
			return err
		}
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
		return err
	}

	destinationFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	zippedFile, err := f.Open()
	if err != nil {
		return err
	}
	defer zippedFile.Close()

	if _, err := io.Copy(destinationFile, zippedFile); err != nil {
		return err
	}
	return nil
}
