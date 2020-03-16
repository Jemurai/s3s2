package main_test

import (
    "runtime"
	"testing"
	file "github.com/tempuslabs/s3s2/file"
	"github.com/stretchr/testify/assert"

)


func TestGetSourceName(t * testing.T) {
    assert := assert.New(t)

    input_file_path := "batch_folder_timestamp/image_001.pdf"
    file_struct := file.File{Name: input_file_path}

    actual := file_struct.GetSourceName("download_directory")

    var expected string
    if runtime.GOOS == "windows" {
        expected = "download_directory\\batch_folder_timestamp\\image_001.pdf"

    } else {
        expected = "download_directory/batch_folder_timestamp/image_001.pdf"
    }
    assert.Equal(actual, expected)
}

func TestGetZipName(t * testing.T) {
    assert := assert.New(t)

    input_file_path := "batch_folder_timestamp/image_001.pdf"
    file_struct := file.File{Name: input_file_path}

    actual := file_struct.GetZipName("download_directory")

    var expected string
    if runtime.GOOS == "windows" {
        expected = "download_directory\\batch_folder_timestamp\\image_001.pdf.zip"

    } else {
        expected = "download_directory/batch_folder_timestamp/image_001.pdf.zip"
    }
    assert.Equal(actual, expected)
}

func TestGetEncryptedName(t * testing.T) {
    assert := assert.New(t)

    input_file_path := "batch_folder_timestamp/image_001.pdf"
    file_struct := file.File{Name: input_file_path}

    actual := file_struct.GetEncryptedName("download_directory")

    var expected string
    if runtime.GOOS == "windows" {
        expected = "download_directory\\batch_folder_timestamp\\image_001.pdf.zip.gpg"

    } else {
        expected = "download_directory/batch_folder_timestamp/image_001.pdf.zip.gpg"
    }
    assert.Equal(actual, expected)
}

func TestGetSourceNameSingleDir(t * testing.T) {
    assert := assert.New(t)

    input_file_path := "image_001.pdf"
    file_struct := file.File{Name: input_file_path}

    actual := file_struct.GetSourceName("download_directory")

    var expected string
    if runtime.GOOS == "windows" {
        expected = "download_directory\\image_001.pdf"

    } else {
        expected = "download_directory/image_001.pdf"
    }
    assert.Equal(actual, expected)
}

func TestGetZipNameSingleDir(t * testing.T) {
    assert := assert.New(t)

    input_file_path := "image_001.pdf"
    file_struct := file.File{Name: input_file_path}

    actual := file_struct.GetZipName("download_directory")

    var expected string
    if runtime.GOOS == "windows" {
        expected = "download_directory\\image_001.pdf.zip"

    } else {
        expected = "download_directory/image_001.pdf.zip"
    }
    assert.Equal(actual, expected)
}

func TestGetEncryptedNameSingleDir(t * testing.T) {
    assert := assert.New(t)

    input_file_path := "image_001.pdf"
    file_struct := file.File{Name: input_file_path}

    actual := file_struct.GetEncryptedName("download_directory")

    var expected string
    if runtime.GOOS == "windows" {
        expected = "download_directory\\image_001.pdf.zip.gpg"

    } else {
        expected = "download_directory/image_001.pdf.zip.gpg"
    }
    assert.Equal(actual, expected)
}


func TestFileStructMethods(t *testing.T) {
    TestGetSourceName(t)
    TestGetZipName(t)
    TestGetEncryptedName(t)

    TestGetSourceNameSingleDir(t)
    TestGetZipNameSingleDir(t)
    TestGetEncryptedNameSingleDir(t)


}
