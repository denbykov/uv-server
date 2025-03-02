package business

type DatabaseFacade interface {
	GetVersion() (string, error)
	UpdateVersion() error
}
