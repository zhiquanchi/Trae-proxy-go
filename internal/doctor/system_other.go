//go:build !windows

package doctor

type SystemProxy struct {
	Enabled  bool
	Server   string
	Override string
	Source   string
}

func DetectSystemProxy() *SystemProxy {
	return nil
}
