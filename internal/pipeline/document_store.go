package pipeline

type DocumentStore interface {
	CreateCollection(name string, dimension int) error
	Exists(collection string, id uint64) (bool, error)
	AddDocument(collection string, id uint64, text string, embedding []float32) error
	Search(collection string, embedding []float32) (string, error)
}
