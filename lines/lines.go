package lines

type Repo interface {
	GetLinesFor(id string) ([]string, error)
}
