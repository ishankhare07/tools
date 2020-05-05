package output

import (
	"fmt"
	"os"

	k8s "istio.io/tools/isotope/convert/pkg/kubernetes"
)

func CreateAndPopulateFilesInDirectory(dir string, manifestMap k8s.ManifestMap) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.Mkdir(dir, os.ModeDir|0777)
		if err != nil {
			return err
		}
	}

	for fileName, fileContent := range manifestMap {
		err := CreateAndPopulateFile(fmt.Sprintf("%s/%s.yaml", dir, fileName), fileContent)

		if err != nil {
			return err
		}
	}

	return nil
}

func CreateAndPopulateFile(fileName string, fileContent string) error {
	f, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer f.Close()

	f.WriteString(fileContent)
	return nil
}