package file

import (
    "strings"
    "path/filepath"
    "os"

    "github.com/karrick/godirwalk"

    options "github.com/tempuslabs/s3s2_new/options"
	utils "github.com/tempuslabs/s3s2_new/utils"
)


type File struct {
	OsRelPath string
	AwsRelPath string
	// additional attributes as needed
}

func (f *File) GetCleanedPath() string {
    return utils.ToSlashClean(f.OsRelPath)
}

func (f *File) GetPosixPath() string {
    return utils.ToPosixPath(f.GetCleanedPath())
}

func (f *File) GetRelativePath(relative_to string) string {
    return utils.GetRelativePath(f.OsRelPath, relative_to)
}

// Specify the target path of the file as it will exist in the organization's folder
func (f *File) SetTargetAwsKey(relative_to string) string {
    return utils.GetRelativePath(f.GetEncryptedName(), relative_to)
}

// Specify the naming convention of the zipped version of the file
func (f *File) GetZipName() string {
    return f.OsRelPath + ".zip"
}

// Specify the naming convention of the encrypted version of the file
func (f *File) GetEncryptedName() string {
    return f.GetZipName() + ".gpg"
}

// Return os.File of the encrypted version of the file
func (f *File) OpenEncryptedFile() *os.File {
    encrypted_file, err := os.Open(f.GetEncryptedName())
    if err != nil {
        panic(err)
    }
    return encrypted_file
}

// Traverse input directory and instantiate File structs
// Any filtering of which types of files to upload can be done here
func GetFileStructsFromDir(directory string, opts options.Options) ([]File, error) {
	var file_structs []File

	err := godirwalk.Walk(directory, &godirwalk.Options{
	        Callback: func(osPathname string, de *godirwalk.Dirent) error {

	            basename := filepath.Base(osPathname)

	            //exclusion criteria
	            not_dir := !de.IsDir()
	            not_manifest := !strings.HasSuffix(basename, "manifest.json")
	            not_zip := !strings.HasSuffix(basename, ".zip")
	            not_gpg := !strings.HasSuffix(basename, ".zip.gpg")
	            not_private := !strings.HasPrefix(basename, ".")

                if not_dir && not_manifest && not_zip && not_gpg && not_private{
                    fs := File{OsRelPath: osPathname, AwsRelPath: utils.GetRelativePath(osPathname, opts.Directory)}
                    file_structs = append(file_structs, fs)
                }
                return nil
                },
            Unsorted: true, // (optional) set true for faster yet non-deterministic enumeration (see godoc)
        })

    return file_structs, err
}