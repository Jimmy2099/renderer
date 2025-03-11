package generate_train_dataset

import (
	_ "embed"
)

//go:embed test.py
var PythonScript string
