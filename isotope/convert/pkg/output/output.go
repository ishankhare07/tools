package output

import (
	"fmt"
	"os"

	k8s "istio.io/tools/isotope/convert/pkg/kubernetes"
)

func CreateAndPopulateFiles(dir string, manifestMap k8s.ManifestMap) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.Mkdir(dir, os.ModeDir|0777)
		if err != nil {
			return err
		}
	}

	for k, v := range manifestMap {
		err := func(k string, v string) error {
			f, err := os.Create(fmt.Sprintf("%s/%s.yaml", dir, k))
			if err != nil {
				return err
			}
			defer f.Close()

			f.WriteString(v)
			return nil
		}(k, v)

		if err != nil {
			return err
		}
	}

	return nil
}
