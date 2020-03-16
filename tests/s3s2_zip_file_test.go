package main_test

import (
    "runtime"
	"testing"

    "io"
	"os"

	zip "github.com/tempuslabs/s3s2_new/zip"

	"github.com/stretchr/testify/assert"

)

func fileExists(filename string) bool {
    _, err := os.Stat(filename)
    if os.IsNotExist(err) {
        return false
    }
    return true
}

func writeToFile(filename string, data string) error {
    file, err := os.Create(filename)
    if err != nil {
        return err
    }
    defer file.Close()

    _, err = io.WriteString(file, data)
    if err != nil {
        return err
    }
    return file.Sync()
}

func setUpEnv() (string, string) {
    var input_file_path string
    var output_file_path string

    if runtime.GOOS == "windows" {
        input_file_path = "s3s2_test_zip_file_creation\\image_001.pdf"
        output_file_path = "s3s2_test_zip_file_creation\\image_001.pdf.zip"

    } else {
        input_file_path = "s3s2_test_zip_file_creation/image_001.pdf"
        output_file_path = "s3s2_test_zip_file_creation/image_001.pdf.zip"
    }
    return input_file_path, output_file_path
}

// Test that given a clean environment the Zipfile function creates a zip file of the appropriate name
func TestZipFileCreation(t * testing.T) {
    assert := assert.New(t)

    input_file_path, output_file_path := setUpEnv()

    writeToFile(input_file_path, "This is test data")
    assert.True(fileExists(input_file_path))

    zip.ZipFile(input_file_path, output_file_path, "")
    assert.True(fileExists(output_file_path))
}

func TestUnZipFileCreation(t * testing.T) {
    assert := assert.New(t)

    input_file_path, output_file_path := setUpEnv()

    // remove original file so we can check that our function creates it
    assert.True(fileExists(input_file_path))
    os.Remove(input_file_path)
    assert.False(fileExists(input_file_path))

    zip.UnZipFile(output_file_path, input_file_path, "")
    assert.True(fileExists(output_file_path))
}

func TestZipFile(t *testing.T) {
    os.RemoveAll("s3s2_test_zip_file_creation")
    os.MkdirAll("s3s2_test_zip_file_creation", os.ModePerm)

    TestZipFileCreation(t)
    TestUnZipFileCreation(t)
}
