package util

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"regexp"
)

/***************************************************************************************************************/

/* Save file from buffer */
func SaveFileFromBuffer(filePath string, buffer []byte) error {
	/* Create folder to store file */
	err := os.MkdirAll(
		filepath.Dir(filePath),
		os.ModePerm,
	)
	if err != nil {
		return err
	}

	/* Write buffer to new file */
	return os.WriteFile(filePath, buffer, os.ModePerm)
}

/* Save file from reader */
func SaveFileFromReader(filePath string, reader io.Reader) error {
	/* Create folder to store file */
	err := os.MkdirAll(
		filepath.Dir(filePath),
		os.ModePerm,
	)
	if err != nil {
		return err
	}

	/* Create or open new file */
	newFile, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer newFile.Close()

	/* Copy data from reader to new file */
	_, err = io.Copy(newFile, reader)
	if err != nil {
		return err
	}

	return nil
}

/* Copy file*/
func CopyFile(srcPath string, desPath string) error {
	if desPath == srcPath {
		return nil
	}

	/* Open file */
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return err
	}

	/* Create folder to store file */
	err = os.MkdirAll(
		filepath.Dir(desPath),
		os.ModePerm,
	)
	if err != nil {
		return err
	}

	/* Create or open new file */
	desFile, err := os.Create(desPath)
	if err != nil {
		return err
	}
	defer desFile.Close()

	/* Copy data from source file to destination file */
	_, err = io.Copy(desFile, srcFile)
	if err != nil {
		return err
	}

	return nil
}

/* Remove all files or folders */
func RemoveAll(fileOrDirPaths []string) error {
	for idx := range fileOrDirPaths {
		err := os.RemoveAll(fileOrDirPaths[idx])
		if err != nil {
			return err
		}
	}

	return nil
}

/* Create folders */
func CreateFolders(dirPaths []string, isRecreate bool) error {
	if isRecreate {
		err := RemoveAll(dirPaths)
		if err != nil {
			return err
		}
	}

	for idx := range dirPaths {
		err := os.MkdirAll(dirPaths[idx], os.ModePerm)
		if err != nil {
			return err
		}
	}

	return nil
}

/* Unzip zip file */
func Unzip(zipFilePath string, dirPath string) error {
	/* Open zipfile */
	reader, err := zip.OpenReader(zipFilePath)
	if err != nil {
		return err
	}
	defer reader.Close()

	/* Create folder to store unzipfile */
	err = os.MkdirAll(dirPath, os.ModePerm)
	if err != nil {
		return err
	}

	// Iterate over each file in zip file
	for _, file := range reader.File {
		desFilePath := filepath.Join(dirPath, file.Name)

		if file.FileInfo().IsDir() {
			err := os.MkdirAll(desFilePath, os.ModePerm)
			if err != nil {
				return err
			}
		} else {
			err := os.MkdirAll(filepath.Dir(desFilePath), os.ModePerm)
			if err != nil {
				return err
			}

			zippedFile, err := file.Open()
			if err != nil {
				return err
			}
			defer zippedFile.Close()

			newFile, err := os.Create(desFilePath)
			if err != nil {
				return err
			}
			defer newFile.Close()

			_, err = io.Copy(newFile, zippedFile)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

/* Zip all files and folders in dir path */
func Zip(dirPath string, filePath string) error {
	/* Create folder to store zip file */
	err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm)
	if err != nil {
		return err
	}

	/* Create zip */
	zipFile, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	/* Create zip writer */
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	return filepath.Walk(
		dirPath,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			header, err := zip.FileInfoHeader(info)
			if err != nil {
				return err
			}

			header.Name, err = filepath.Rel(dirPath, path)
			if err != nil {
				return err
			}

			if info.IsDir() {
				header.Name += "/"

				return nil
			}

			writer, err := zipWriter.CreateHeader(header)
			if err != nil {
				return err
			}

			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			_, err = io.Copy(writer, file)
			if err != nil {
				return err
			}

			return nil
		},
	)
}

/* Zip files */
func ZipFiles(iFilePaths []string, oFilePath string) error {
	/* Create folder to store zip file */
	err := os.MkdirAll(filepath.Dir(oFilePath), os.ModePerm)
	if err != nil {
		return err
	}

	/* Create zip */
	zipFile, err := os.Create(oFilePath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	/* Create zip writer */
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	for idx := range iFilePaths {
		err := filepath.Walk(
			iFilePaths[idx],
			func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				header, err := zip.FileInfoHeader(info)
				if err != nil {
					return err
				}

				header.Name, err = filepath.Rel(filepath.Dir(iFilePaths[idx]), path)
				if err != nil {
					return err
				}

				if info.IsDir() {
					header.Name += "/"

					return nil
				}

				writer, err := zipWriter.CreateHeader(header)
				if err != nil {
					return err
				}

				file, err := os.Open(path)
				if err != nil {
					return err
				}
				defer file.Close()

				_, err = io.Copy(writer, file)
				if err != nil {
					return err
				}

				return nil
			},
		)
		if err != nil {
			return err
		}
	}

	return nil
}

/* Search files in folder */
func SearchFiles(dirPath string, pattern string, recurse bool, includeDirPath bool) ([]string, error) {
	results := []string{}

	regex, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		entryName := entry.Name()

		if entry.IsDir() {
			if recurse {
				subResults, err := SearchFiles(filepath.Join(dirPath, entryName), pattern, recurse, includeDirPath)
				if err != nil {
					return nil, err
				}

				results = append(results, subResults...)
			}
		} else if regex.MatchString(entryName) {
			if includeDirPath {
				results = append(results, filepath.Join(dirPath, entryName))
			} else {
				results = append(results, entryName)
			}
		}
	}

	return results, nil
}

/* Search dirs in folder */
func SearchDirs(dirPath string, pattern string, recurse bool, includeDirPath bool) ([]string, error) {
	results := []string{}

	regex, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		entryName := entry.Name()

		if entry.IsDir() {
			if regex.MatchString(entryName) {
				if includeDirPath {
					results = append(results, filepath.Join(dirPath, entryName))
				} else {
					results = append(results, entryName)
				}
			}

			if recurse {
				subResults, err := SearchDirs(filepath.Join(dirPath, entryName), pattern, recurse, includeDirPath)
				if err != nil {
					return nil, err
				}

				results = append(results, subResults...)
			}
		}
	}

	return results, nil
}

/* Check file is valid? */
func IsValidFile(filePath string) bool {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return false
	}

	return !fileInfo.IsDir()
}

/* Check directory is valid? */
func IsValidDir(dirPath string) bool {
	dirInfo, err := os.Stat(dirPath)
	if err != nil {
		return false
	}

	return dirInfo.IsDir()
}

/* Get filename without extension */
func GetFileNameWoExt(fileName string) string {
	return fileName[:len(fileName)-len(filepath.Ext(fileName))]
}

/***************************************************************************************************************/
