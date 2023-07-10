package bundle

type Viewer interface {
	Info() Info
	String() string
	Close() error
}
