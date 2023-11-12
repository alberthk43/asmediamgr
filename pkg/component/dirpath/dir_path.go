package dirpath

type DirPath struct {
	configPath   string
	motherPath   string
	targetTvPath string
}

type PathOption func(*DirPath)

func ConfigPathOption(configPath string) PathOption {
	return func(d *DirPath) {
		d.configPath = configPath
	}
}

func MotherPathOption(motherPath string) PathOption {
	return func(d *DirPath) {
		d.motherPath = motherPath
	}
}

func TargetTvPathOption(targetPath string) PathOption {
	return func(d *DirPath) {
		d.targetTvPath = targetPath
	}
}

func NewDirPath(opts ...PathOption) (*DirPath, error) {
	d := &DirPath{
		configPath:   ".",
		motherPath:   ".",
		targetTvPath: ".",
	}
	for _, opt := range opts {
		opt(d)
	}
	return d, nil
}

func (d *DirPath) ConfigPath() string {
	return d.configPath
}

func (d *DirPath) MotherPath() string {
	return d.motherPath
}

func (d *DirPath) TargetTvPath() string {
	return d.targetTvPath
}
