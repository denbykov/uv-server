package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	fmt.Println("starting...")

	serverTagsUrl := "https://api.github.com/repos/denbykov/uv-server/tags"
	serverVersion, err := getLatestVersion(serverTagsUrl)
	if err != nil {
		fmt.Printf("failed to get latest server version: %v\n", err)
		return
	}
	fmt.Printf("latest server version: %v\n", serverVersion)

	clientTagsUrl := "https://api.github.com/repos/denbykov/uv-client/tags"
	clientVersion, err := getLatestVersion(clientTagsUrl)
	if err != nil {
		fmt.Printf("failed to get latest client version: %v\n", err)
		return
	}
	fmt.Printf("latest client version: %v\n", clientVersion)

	err = installServer(serverVersion)
	if err != nil {
		fmt.Printf("failed to install server: %v\n", err)
		return
	}

	err = installClient(clientVersion)
	if err != nil {
		fmt.Printf("failed to install client: %v\n", err)
		return
	}

	fmt.Printf("update completed")
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

	err = json.Unmarshal([]byte(buf.String()), &tags)

	return tags, err
}

func installServer(tag string) error {
	url := fmt.Sprintf("https://github.com/denbykov/uv-server/releases/download/%v/server.zip", tag)
	err := downloadFile(url, "server.zip")
	if err != nil {
		return err
	}

	err = unzipSource("server.zip", "server")
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

	err = unzipSource("client.zip", "client")
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

	_, err = io.Copy(out, resp.Body)
	return err
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
