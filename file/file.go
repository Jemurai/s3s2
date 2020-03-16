package file

import (
    "strings"
    "path/filepath"

    "github.com/karrick/godirwalk"

    options "github.com/tempuslabs/s3s2_new/options"
	utils "github.com/tempuslabs/s3s2_new/utils"

)


type File struct {
	Name string
	// additional attributes as needed
}

// Specify the naming convention of the original version of the file
func (f *File) GetSourceName(directory string) string {
    return filepath.Join(directory, f.Name)
}

// Specify the naming convention of the zipped version of the file
func (f *File) GetZipName(directory string) string {
    return filepath.Join(directory, f.Name + ".zip")
}

// Specify the naming convention of the encrypted version of the file
func (f *File) GetEncryptedName(directory string) string {
    return filepath.Join(directory, f.Name + ".zip.gpg")
}

// Traverse input directory and instantiate File structs
// Any filtering of which types of files to upload can be done here
func GetFileStructsFromDir(directory string, opts options.Options) ([]File, error) {
	var file_structs []File

	err := godirwalk.Walk(directory, &godirwalk.Options{
	        Callback: func(file_path string, de *godirwalk.Dirent) error {

	            basename := filepath.Base(file_path)

	            //exclusion criteria
	            not_dir := !de.IsDir()
	            not_manifest := !strings.HasSuffix(basename, "manifest.json")
	            not_private := !strings.HasPrefix(basename, ".")
	            // in event of "dirty" run, don't upload existing zip or gpg files
	            not_zip := !strings.HasSuffix(basename, ".zip")
	            not_gpg := !strings.HasSuffix(basename, ".zip.gpg")

                if not_dir && not_manifest && not_zip && not_gpg && not_private {
                    file_path := utils.GetRelativePath(file_path, opts.Directory)
                    file_structs = append(file_structs, File{Name: file_path})
                }
                return nil
                },
            Unsorted: false, // (optional) set true for faster yet non-deterministic enumeration (see godoc)
        })

    return file_structs, err
}
