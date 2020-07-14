package file

import (
    "errors"
    "strings"
    "os"
    "fmt"
    "path/filepath"

    godirwalk "github.com/karrick/godirwalk"
    log "github.com/sirupsen/logrus"

    options "github.com/tempuslabs/s3s2/options"
	utils "github.com/tempuslabs/s3s2/utils"

)

type File struct {
	Name string
	// additional attributes as needed
}

// Specify the filepath of the original version of the file
func (f *File) GetSourceName(directory string) string {
    return filepath.Join(directory, f.Name)
}

// Specify the filepath of the zipped version of the file
func (f *File) GetZipName(directory string) string {
    return filepath.Join(directory, f.Name + ".zip")
}

// Specify the filepath of the encrypted version of the file
func (f *File) GetEncryptedName(directory string) string {
    return filepath.Join(directory, f.Name + ".zip.gpg")
}

// Break an array of objects into an array of chunks of n size
func ChunkArray(in_array []File, chunk_size int) [][]File {

   var chunks [][]File

   for i := 0; i < len(in_array); i += chunk_size {
        end := i + chunk_size

        if end > len(in_array) {
            end = len(in_array)
        }
        chunks = append(chunks, in_array[i:end])
    }
    return chunks
}

// function housing the logic to determine what files are instantiated and eventually processed.
// for example used to ignore private files that may be filesystem-specific (i.e. .nfs files)
// or encrypted and zipped files from a previous run
func includeFile(de *godirwalk.Dirent, basename string, opts options.Options) bool {

    not_dir := !de.IsDir()
    not_manifest := !strings.HasSuffix(basename, "manifest.json")
    not_private := !strings.HasPrefix(basename, ".")
    not_zip := !strings.HasSuffix(basename, ".zip")
    not_gpg := !strings.HasSuffix(basename, ".zip.gpg")

    if not_dir && not_manifest && not_private && not_zip && not_gpg {
        return true
    } else {
        return false
    }
}

// Traverse input directory and instantiate File structs
// Will split metadata from other file structs
func GetFileStructsFromDir(directory string, opts options.Options) ([]File, []File, error) {
	var file_structs_metadata []File
	var file_structs []File

	err := godirwalk.Walk(directory, &godirwalk.Options{
	        Callback: func(file_path string, de *godirwalk.Dirent) error {
	            log.Debugf("Walking: '%s'", file_path)
	            basename := filepath.Base(file_path)
                if includeFile(de, basename, opts) {
                    log.Debugf("Registering '%s' to manifest", file_path)
                    file_path, err := filepath.Rel(opts.Directory, file_path)
                    utils.PanicIfError("Unable to discern relative path - ", err)
                    // if current file is a metadata file, append to dedicated metadata chunk
                    if utils.Include(opts.MetaDataFiles, basename) {
                    	file_structs_metadata = append(file_structs_metadata, File{Name: file_path})
                    // otherwise append to normal file chunk
                    } else {
                        file_structs = append(file_structs, File{Name: file_path})
                    }
                } else {
                    log.Debugf("Skipping over file '%s' - this file will NOT be encrypted...", file_path)
                }
                return nil
                },
            ErrorCallback: func(osPathname string, err error) godirwalk.ErrorAction {
           	// Your program may want to log the error somehow.
           	log.Errorf(os.Stderr, "Error walking input directory: %s\n", err),
            Unsorted: false, // (optional) set true for faster yet non-deterministic enumeration (see godoc)
        })
    // if we expect metadata files and don't pick them up, there might be a typo
    if len(opts.MetaDataFiles) > 0 && len(file_structs_metadata) == 0 {
        err = errors.New("Metadata files specified but none identified in input directory. Check your spelling and that the files exist in the input directory")
        panic(err)
    }

    log.Infof("Identified metadata-files '%s'...", file_structs_metadata)

    return file_structs, file_structs_metadata, err
}


func ArchiveFileStructs(file_structs_to_archive []File, input_dir string, archive_dir string) error {

    os.MkdirAll(archive_dir, os.ModePerm)

    var source_path string
    var err error
    var dir string

    for _, fs := range file_structs_to_archive {
        source_path = fs.GetSourceName(input_dir)

        utils.PanicIfError(fmt.Sprintf("Unable to get relative path for '%s'", fs.Name), err)

        archive_full_path := filepath.Join(archive_dir, source_path)
        archive_full_path_dir := filepath.Dir(archive_full_path)

        os.MkdirAll(archive_full_path_dir, os.ModePerm)

        log.Debugf("Archiving file '%s' to '%s'", fs.Name, archive_full_path)
        err = os.Rename(source_path, archive_full_path)

        if err != nil {
            panic(err)
        }

        dir, _ = filepath.Split(filepath.Join(input_dir, source_path))
        is_dir_empty, _ := utils.IsDirEmpty(dir)

        if is_dir_empty {
            os.Remove(dir)
        }
    }
    return err
}

