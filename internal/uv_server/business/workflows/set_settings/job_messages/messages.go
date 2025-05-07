package jobmessages

type Settings struct {
	StorageDir string `json:"storage_dir"`
}

type Request struct {
	Settings Settings `json:"settings"`
}

type Result struct {
	Settings Settings `json:"settings"`
}
