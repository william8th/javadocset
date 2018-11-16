package main

import (
	"os"
	"path"
	"github.com/inconshreveable/log15"
	"path/filepath"
	"errors"
	"strings"
	"io"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"fmt"
)

const OVERVIEW_SUMMARY = "overview-summary.html"

var log = log15.New()

var toIndex []string

func main() {

	arguments := os.Args
	argLength := len(arguments)
	if (argLength == 2 && arguments[1] == "--help") {
		printUsage()
		return
	} else if (argLength != 3) {
		log.Error("Invalid argument(s) provided")
		printUsage()
		os.Exit(1)
	}

	docsetName := path.Clean(arguments[1])
	var javadocPath = path.Clean(arguments[2])

	log.Info("Running with arguments", "docsetName", docsetName, "javadocPath", javadocPath)

	docsetDirectoryPath := docsetName + ".docset"

	if exists, _ := pathExists(docsetDirectoryPath); exists {
		log.Info("Removing existing docset directory", "Docset directory path", docsetDirectoryPath)
		if err := os.RemoveAll(docsetDirectoryPath); err != nil {
			log.Error(
				"Unable to remove existing docset directory",
				"Docset directory path", docsetDirectoryPath,
				"error", err,
			)
			os.Exit(1)
		}
	}

	contentsDirectoryPath := docsetDirectoryPath + "/Contents"
	resourcesDirectoryPath := contentsDirectoryPath + "/Resources"
	documentsDirectoryPath := resourcesDirectoryPath + "/Documents"

	log.Info("Creating docset folder structure...")
	if err := os.MkdirAll(documentsDirectoryPath, os.ModePerm); err != nil {
		log.Error("Unable to create docset folder structure", "Docset directory", docsetDirectoryPath)
		os.Exit(1)
	}

	var docsetIndexFile string
	overviewSummaryPath := javadocPath + OVERVIEW_SUMMARY
	var summaryFound = false

	if exists, _ := pathExists(overviewSummaryPath); !exists {

		walkCount := 0
		filepath.Walk(javadocPath, func(filePath string, info os.FileInfo, err error) error {

			if err != nil {
				log.Error("Failed to walk path", "path", filePath, "err", err)
				os.Exit(1)
			}

			walkCount++

			if walkCount < 10000 {

				if info.Name() == OVERVIEW_SUMMARY {
					javadocPath = path.Dir(filePath)
					summaryFound = true
				}

				return nil
			} else {
				return errors.New("Hit file enumeration limit")
			}
		})
	} else {
		summaryFound = true
	}

	if summaryFound {
		docsetIndexFile = OVERVIEW_SUMMARY
	}

	hasMultipleIndices := false
	indexFilesPath := javadocPath + "index-files"
	if exists, _ := pathExists(indexFilesPath); exists {
		if docsetIndexFile == "" {
			docsetIndexFile = "index-files/index-1.html"
		}
		hasMultipleIndices = true
	}
	log.Info("Done!")

	copyFiles(documentsDirectoryPath, javadocPath)

	documentsDirectoryIndex := documentsDirectoryPath + "/index-all.html"
	if exists, _ := pathExists(documentsDirectoryIndex); !hasMultipleIndices && exists {

		toIndex = append(toIndex, documentsDirectoryIndex)

		if docsetIndexFile == "" {
			docsetIndexFile = "index-all.html"
		}

	} else {

		indexFilesPath := documentsDirectoryPath + "/index-files"
		filepath.Walk(indexFilesPath, func(filePath string, info os.FileInfo, err error) error {

			if err != nil {
				log.Error("Failed to walk path", "filePath", filePath, "err", err)
				os.Exit(1)
			}

			filename := info.Name()
			if strings.HasPrefix(filename, "index-") && strings.HasSuffix(filename, ".html") {
				toIndex = append(toIndex, filePath)
			}
			return err
		})

	}

	if len(toIndex) == 0 {
		log.Error("API folder specified does not contain any index files (either an 'index-all.html' file or an 'index-files' folder and is not valid")
		printUsage()
		return
	}

	writeInfoPlist(docsetName, docsetIndexFile, contentsDirectoryPath)

	initDB(resourcesDirectoryPath, index(toIndex))
}

func printUsage() {
	log.Info("Usage: javadocset <docset name> <javadoc API folder>")
	log.Info("<docset name> - anything you want")
	log.Info("<javadoc API folder> - the path of the javadoc API folder you want to index")
}

func copyFiles(documentsDirectoryPath, javadocPath string) {
	log.Info("Copying files...", "source", javadocPath, "destination", documentsDirectoryPath)

	src := path.Clean(javadocPath)
	dst := path.Clean(documentsDirectoryPath)

	srcBase := path.Base(src)
	filepath.Walk(src, func(filePath string, info os.FileInfo, err error) error {

		if err != nil {
			log.Error("Error walking path", "filePath", filePath)
			os.Exit(1)
		}

		if info.IsDir() {

			if path.Base(filePath) != srcBase {

				// We only want to copy the directories within the source directory
				// to the destination directory
				directoryName := strings.Split(filePath, srcBase)[1]

				err := os.MkdirAll(dst + directoryName, os.ModePerm)

				if err != nil {
					log.Error("Unable to create directory", "directory", directoryName)
					os.Exit(1)
				}
			}

		} else {

			// Copy file
			fileName := filepath.Base(filePath)
			directoryName := strings.Split(filepath.Dir(filePath), srcBase)[1]

			dstPath := filepath.Clean(dst + directoryName + "/" + fileName)

			err = copyFileContents(filePath, dstPath)

			if err != nil {
				log.Error("Unable to copy file", "src", filePath, "dst", dstPath)
				os.Exit(1)
			}
		}

		return err
	})

	log.Info("Done!")
}

func writeInfoPlist(docsetName, docsetIndexFile, contentsDirectoryPath string) {
	plistContentTemplate := "<?xml version=\"1.0\" encoding=\"UTF-8\"?><plist version=\"1.0\"><dict><key>CFBundleIdentifier</key> <string>%v</string><key>CFBundleName</key> <string>%v</string> <key>DocSetPlatformFamily</key> <string>%v</string> <key>dashIndexFilePath</key><string>%v</string><key>DashDocSetFamily</key><string>java</string><key>isDashDocset</key><true/></dict></plist>"

	docsetIdentifier := firstPhraseLowerCased(docsetName)

	plistContent := fmt.Sprintf(
		plistContentTemplate,
		docsetIdentifier,
		docsetName,
		docsetIdentifier,
		docsetIndexFile,
	)

	infoPlistPath := contentsDirectoryPath + "/Info.plist"
	err := writeStringToFile(plistContent, infoPlistPath)
	if err != nil {
		log.Error("Unable to write to plist file", "plistPath", infoPlistPath)
	}
}

func initDB(resourcesDirectoryPath string, dbFunc func(*sql.DB)) {

	dbPath := filepath.Clean(resourcesDirectoryPath + "/docSet.dsidx")

	// We don't care, we just want to remove the index
	os.Remove(dbPath)

	db, err := sql.Open("sqlite3", dbPath)

	if err != nil {
		log.Error("Unable to create sqlite database", "destination", dbPath, "error", err)
		os.Exit(1)
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE searchIndex(id INTEGER PRIMARY KEY, name TEXT, type TEXT, path TEXT)")
	if err != nil {
		log.Error("Unable to create table", "error", err)
		os.Exit(1)
	}

	if dbFunc != nil {
		dbFunc(db)
	}
}

func index(indicesToIndex []string) func(db *sql.DB) {
	return func(db *sql.DB) {

		tx, err := db.Begin()
		if err != nil {
			log.Error("Unable to begin transactions for database", "error", err)
			os.Exit(1)
		}

		stmt, err := tx.Prepare("INSERT INTO searchIndex(name, type, path) VALUES (?, ?, ?)")
		if err != nil {
			log.Error("Unable to create statement to insert into database", "error", err)
			os.Exit(1)
		}
		defer stmt.Close()

		added := make(map[string]bool)

		for _, toIndex := range indicesToIndex {
			parseIndex(toIndex, func(entry IndexEntry) {

				name, elementType, path := entry.name, entry.elementType.value(), entry.path

				uniqueKey := name + elementType + path

				if !added[uniqueKey] {
					_, err := stmt.Exec(name, elementType, path)
					if err != nil {
						log.Error(
							"Unable to insert entry",
							"name", name,
							"elementType", elementType,
							"path", path,
						)
						os.Exit(1)
					}

					added[uniqueKey] = true
				}
			})
		}

		tx.Commit()
	}
}


/**
Utility functions
 */
func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func writeStringToFile(content, dst string) error {
	file, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write([]byte(content))
	return err
}

func copyFileContents(src, dst string) (err error) {

	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		log.Error("Error copying", "error", err)
		return
	}
	err = out.Sync()
	return
}

func firstPhraseLowerCased(s string) string {
	return strings.ToLower(func() string {
		return strings.Split(s, " ")[0]
	}())
}
