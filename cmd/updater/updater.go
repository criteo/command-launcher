package updater

type Updater interface {
	CheckUpdateAsync()
	Update()
}
